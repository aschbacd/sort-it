package utils

import (
	"os"
	"os/exec"
)

// CheckFolder checks whether a given directory exists or not
func CheckFolder(path string) bool {
	if _, err := os.Stat(path); err != nil {
		// Invalid folder
		return false
	}

	return true
}

// CommandAvailable checks whether a command is available or not
func CommandAvailable(command string) bool {
	if _, err := exec.LookPath(command); err != nil {
		return false
	}

	return true
}
