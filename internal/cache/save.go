package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

func Save(store Store) error {
	path, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(store); err != nil {
		f.Close()
		return err
	}
	f.Close()

	return os.Rename(tmp, path)
}
