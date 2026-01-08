package cache

import (
	"encoding/json"
	"os"
)

func Load() (Store, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Store{}, nil // empty cache
		}
		return nil, err
	}
	defer f.Close()

	var store Store
	if err := json.NewDecoder(f).Decode(&store); err != nil {
		return nil, err
	}

	return store, nil
}
