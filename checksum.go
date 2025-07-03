package main

import (
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"io"
	"os"
	"strings"
)

func checksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}
	defer file.Close()

	hasher := sha1.New()
	buf := make([]byte, 64*1024) // 64 KiB buffer
	_, err = io.CopyBuffer(hasher, file, buf)
	if err != nil {
		return "", fmt.Errorf("failed to hash: %v", err)
	}

	encoded := base32.StdEncoding.EncodeToString(hasher.Sum(nil))
	return strings.TrimRight(encoded, "="), nil
}