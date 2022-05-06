package utils

import (
	"os/exec"
)

// CommandAvailable checks whether a command is available or not
func CommandAvailable(command string) bool {
	if _, err := exec.LookPath(command); err != nil {
		return false
	}

	return true
}

// Contains checks if a string slice contains a string
func Contains(slice []string, str string) bool {
	for _, item := range slice {
		if item == str {
			return true
		}
	}
	return false
}
