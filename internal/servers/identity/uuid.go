package identity

import (
	"os"
	"path/filepath"
	"renderctl/logger"
	"strings"

	"github.com/google/uuid"
)

func loadOrCreateUUID(path string) (string, error) {
	// ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return "", err
	}

	b, err := os.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(b)), nil
	}

	id := uuid.New().String()
	if err := os.WriteFile(path, []byte(id), 0600); err != nil {
		return "", err
	}

	return id, nil
}

func FetchUUID() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	path := filepath.Join(home, ".renderctl", "server_uuid")
	logger.Info("UUID path: %s", path)
	return loadOrCreateUUID(path)
}
