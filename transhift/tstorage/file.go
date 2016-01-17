package tstorage

import (
	"os"
	"crypto/sha256"
	"io"
)

func HashFile(file *os.File) ([]byte, error) {
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}
