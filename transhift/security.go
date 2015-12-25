package transhift

import (
    "crypto/sha256"
    "os"
    "io"
)

func calculateFileChecksum(file *os.File) []byte {
    hash := sha256.New()
    io.Copy(hash, file)
    return hash.Sum(nil)
}
