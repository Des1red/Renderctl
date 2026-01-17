package cache

import (
	"os"
	"path/filepath"
)

func Path() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".renderctl", "devices.json"), nil
}
