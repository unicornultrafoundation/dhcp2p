package ethereum

import (
	"fmt"
	"os"
	"strings"
)

// readPasswordFromFile reads password from a file if the path is provided
func readPasswordFromFile(passwordPath string) (string, error) {
	if passwordPath == "" {
		return "", nil
	}

	// Check if it's a file path
	if _, err := os.Stat(passwordPath); err == nil {
		// File exists, read password from it
		passwordBytes, err := os.ReadFile(passwordPath)
		if err != nil {
			return "", fmt.Errorf("failed to read password file %s: %w", passwordPath, err)
		}
		// Trim whitespace and newlines from password
		return strings.TrimSpace(string(passwordBytes)), nil
	}

	// Not a file path, return as is
	return passwordPath, nil
}
