package utils

import (
	"io"
	"os"
)

// DirectoryExists checks if a directory path exists
func DirectoryExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Does not exist
		return false
	}

	return true
}

// DirectoryIsEmpty checks if a directory is empty
func DirectoryIsEmpty(path string) (bool, error) {
	// Open directory
	dir, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// Check if directory has at least 1 record
	if _, err = dir.ReadDir(1); err == io.EOF {
		return true, nil
	}

	return false, nil
}

// SliceContainsString checks if a string slice contains a string
func SliceContainsString(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
