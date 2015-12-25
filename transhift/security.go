package transhift

import (
    "crypto/sha256"
    "os"
    "io"
)

func calculateStringChecksum(x string) []byte {
    hash := sha256.New()
    hash.Write([]byte(x))
    return hash.Sum(nil)
}

func calculateFileChecksum(file *os.File) []byte {
    hash := sha256.New()
    io.Copy(hash, file)
    return hash.Sum(nil)
}
