package utils

import (
	"os"
	"path/filepath"
)

func GetSVaultDir() (string, error) {
	userDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(userDir, ".svault"), nil
}
