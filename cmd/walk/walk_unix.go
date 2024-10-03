//go:build !plan9 && !windows
package main

import (
	"syscall"
)

// gatherFileInfo collects file information based on the provided path
func gatherFileInfo(path string) (FileInfo, error) {
	var stat syscall.Stat_t
	err := syscall.Stat(path, &stat)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:       path,
		IsDir:      (stat.Mode & syscall.S_IFDIR) != 0,
		IsSymlink:  (stat.Mode & syscall.S_IFLNK) != 0,
	}, nil
}
