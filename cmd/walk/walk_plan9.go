//go:build plan9
package main

import (
	"os"
)

func gatherFileInfo(path string) (FileInfo, error) {
	var fileInfo os.FileInfo
	var err error

	fileInfo, err = os.Stat(path)

	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:       path,
		IsDir:      fileInfo.IsDir(),
		IsSymlink:  fileInfo.Mode()&os.ModeSymlink != 0,
	}, nil
}
