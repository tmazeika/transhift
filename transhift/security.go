package transhift

import (
    "os"
    "crypto/sha256"
    "io"
)

func fileChecksum(file string) ([]byte, error) {
    file, err := os.Open(file)

    if err != nil {
        return nil, err
    }

    defer file.Close()

    hash := sha256.New()
    _, err = io.Copy(hash, file)

    if err != nil {
        return nil, err
    }

    return hash.Sum(nil);
}

func stringChecksum(data string) []byte {
    hash := sha256.New()

    hash.Write([]byte(data))

    return hash.Sum(nil);
}
