package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/vbauerster/mpb/v8"
	"golang.org/x/sync/errgroup"
)

const (
    maxAllowedSize = 10 * 1024 * 1024 * 1024 // 10 GiB
    maxRetries = 3
)

// safeCreateFile ensures destPath is within baseDir and that there is enough free space.
func safeCreateFile(destPath, baseDir string, size int64) (*os.File, error) {
    cleanDest := filepath.Clean(destPath)

    absBase, err := filepath.Abs(baseDir)
    if err != nil {
        return nil, fmt.Errorf("resolving base dir: %w", err)
    }
    absDest, err := filepath.Abs(cleanDest)
    if err != nil {
        return nil, fmt.Errorf("resolving dest path: %w", err)
    }

    rel, err := filepath.Rel(absBase, absDest)
    if err != nil {
        return nil, fmt.Errorf("resolving relative path: %w", err)
    }
    if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
        return nil, fmt.Errorf("destination %q is outside of %q", absDest, absBase)
    }

    var statfs syscall.Statfs_t
    if err := syscall.Statfs(absBase, &statfs); err != nil {
        return nil, fmt.Errorf("statfs on %q: %w", absBase, err)
    }
    freeBytes := int64(statfs.Bavail) * int64(statfs.Bsize)
    if freeBytes < size {
        return nil, fmt.Errorf("not enough disk space: need %d, have %d", size, freeBytes)
    }

    outFile, err := os.Create(absDest)
    if err != nil {
        return nil, fmt.Errorf("creating file: %w", err)
    }
    if err := outFile.Truncate(size); err != nil {
        outFile.Close()
        return nil, fmt.Errorf("preallocating file: %w", err)
    }
    return outFile, nil
}

// DownloadVideo downloads `url` into `destPath` with up to `concurrency` workers.
func DownloadVideo(ctx context.Context, url, destPath, baseDir string, concurrency int, bar *mpb.Bar) error {
    // 1. HEAD to get length and check range support
    req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
    if err != nil {
        return fmt.Errorf("creating HEAD request: %w", err)
    }
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return fmt.Errorf("HEAD request failed: %w", err)
    }
    resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("HEAD returned %s", resp.Status)
    }
    length := resp.ContentLength
    if length <= 0 || length > maxAllowedSize {
        return fmt.Errorf("invalid or too large content length: %d", length)
    }

    acceptRanges := resp.Header.Get("Accept-Ranges")
    supportRanges := acceptRanges == "bytes"

    // 2. Prepare output file (or fall back to streaming if no range support)
    outFile, err := safeCreateFile(destPath, baseDir, length)
    if err != nil {
        return err
    }
    defer outFile.Close()

    // If server doesn't support ranges, just stream it
    if !supportRanges || concurrency <= 1 {
        return singleDownload(ctx, url, outFile, bar)
    }

    // 3. Split into chunks
    partSize := length / int64(concurrency)
    eg, ctx := errgroup.WithContext(ctx)

    for i := 0; i < concurrency; i++ {
        start := int64(i) * partSize
        end := start + partSize - 1
        if i == concurrency-1 {
            end = length - 1
        }

        // capture for closure
        chunkStart, chunkEnd := start, end

        eg.Go(func() error {
            var lastErr error
            for attempt := range maxRetries {
                if err := downloadChunk(ctx, url, outFile, chunkStart, chunkEnd, bar); err != nil {
                    lastErr = err
                    time.Sleep(time.Duration(attempt+1) * 500 * time.Millisecond)
                    continue
                }
                return nil
            }
            return fmt.Errorf("chunk %d-%d failed after %d attempts: %w", chunkStart, chunkEnd, maxRetries, lastErr)
        })
    }

    return eg.Wait()
}

// singleDownload streams the entire file when ranges aren’t supported
func singleDownload(ctx context.Context, url string, outFile *os.File, bar *mpb.Bar) error {
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("download failed: %s", resp.Status)
    }
    reader := bar.ProxyReader(resp.Body)
    _, err = io.Copy(outFile, reader)
    return err
}

// downloadChunk fetches a single byte range [start–end] and writes it at the right offset.
func downloadChunk(ctx context.Context, url string, outFile *os.File, start, end int64, bar *mpb.Bar) error {
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // insist on 206 Partial Content
    if resp.StatusCode != http.StatusPartialContent {
        return fmt.Errorf("expected 206 for range %d-%d, got %d", start, end, resp.StatusCode)
    }
    
    reader := bar.ProxyReader(resp.Body)
    buf := make([]byte, 32*1024)
    offset := start
    for {
        n, readErr := reader.Read(buf)
        if n > 0 {
            if _, writeErr := outFile.WriteAt(buf[:n], offset); writeErr != nil {
                return writeErr
            }
            offset += int64(n)
        }
        if readErr == io.EOF {
            break
        }
        if readErr != nil {
            return readErr
        }
    }
    return nil
}
