package config

import (
	"errors"
	"os"
)

// readFile is a small wrapper that returns (nil, nil) for non-existent files
// so callers can distinguish "absent" from "broken".
func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}
