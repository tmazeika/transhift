package source

import (
	"os"
)

func getFile(path string) (file *os.File, info os.FileInfo, err error) {
	if file, err = os.Open(path); err != nil {
		return
	}

	info, err = file.Stat()
	return
}
