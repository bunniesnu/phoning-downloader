package main

import (
	"crypto/sha1"
	"encoding/base32"
	"fmt"
	"os"
	"strings"
)

func checksum(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("file not found: %s", filePath)
	}

	hash := sha1.Sum(data)
	encoded := base32.StdEncoding.EncodeToString(hash[:])
	return strings.TrimRight(encoded, "="), nil
}