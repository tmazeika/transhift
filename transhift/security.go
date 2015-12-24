package transhift

import (
    "os"
    "crypto/sha256"
    "io"
    "fmt"
)

func fileChecksum(file *os.File) []byte {
    hash := sha256.New()

    if _, err := io.Copy(hash, file); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }

    return hash.Sum(nil);
}

func stringChecksum(data string) []byte {
    hash := sha256.New()

    hash.Write([]byte(data))

    return hash.Sum(nil);
}
