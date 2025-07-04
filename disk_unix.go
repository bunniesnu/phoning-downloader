//go:build !windows
// +build !windows

package main

import (
	"fmt"
	"syscall"
)

func getDiskFreeSpace(path string) (int64, error) {
	var statfs syscall.Statfs_t
    if err := syscall.Statfs(path, &statfs); err != nil {
        return 0, fmt.Errorf("statfs on %q: %w", path, err)
    }
	return int64(statfs.Bavail) * int64(statfs.Bsize), nil
}