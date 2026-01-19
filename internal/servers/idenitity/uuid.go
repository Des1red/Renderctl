package identity

import (
	"os"
	"strings"

	"github.com/google/uuid"
)

func LoadOrCreateUUID(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		return strings.TrimSpace(string(b)), nil
	}

	id := uuid.New().String()
	err = os.WriteFile(path, []byte(id), 0600)
	if err != nil {
		return "", err
	}

	return id, nil
}
