package source

import (
	"os"
	"github.com/transhift/transhift/common/protocol"
	"github.com/transhift/transhift/transhift/storage"
	"path/filepath"
)

func getFile(path string) (file *os.File, info protocol.FileInfo, err error) {
	if file, err = os.Open(path); err != nil {
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	hash, err := storage.HashFile(file)
	if err != nil {
		return
	}

	info = protocol.FileInfo{
		Name: filepath.Base(fileInfo.Name()),
		Size: fileInfo.Size(),
		Hash: hash,
	}
	return
}
