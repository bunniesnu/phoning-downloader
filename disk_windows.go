//go:build windows
// +build windows

package main

import (
	"fmt"

	"golang.org/x/sys/windows"
)

func getDiskFreeSpace(path string) (int64, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, fmt.Errorf("failed to encode path: %v", err)
	}

	var freeBytes, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(pathPtr, &freeBytes, &totalBytes, &totalFreeBytes)
	if err != nil {
		return 0, fmt.Errorf("failed to get disk space: %v", err)
	}

	return int64(freeBytes), nil
}