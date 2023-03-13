// Package helpers implements several useful functions
package helpers

import (
	"fmt"
	"os"
)

// FileExists takes a string returns if it is an existing file
func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// FolderExists takes a string returns if it is an existing folder
func FolderExists(foldername string) bool {
	info, err := os.Stat(foldername)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

// DeleteFile removes the file if it exists
func DeleteFile(filename string) (bool, error) {
	var err error
	if FileExists(filename) {
		err = os.Remove(filename)
		if err != nil {
			return false, err
		}
		return true, nil
	}
	return false, fmt.Errorf("file %s does not exist", filename)
}
