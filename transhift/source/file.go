package source

import (
	"os"
	"github.com/transhift/transhift/common/protocol"
	"github.com/transhift/transhift/transhift/tstorage"
)

func getFile(path string) (file *os.File, info protocol.FileInfo, err error) {
	if file, err = os.Open(path); err != nil {
		return
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return
	}

	info = protocol.FileInfo{
		Name: fileInfo.Name(),
		Size: fileInfo.Size(),
		Hash: tstorage.HashFile(file),
	}
	return
}
