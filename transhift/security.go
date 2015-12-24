package transhift

import (
    "os"
    "crypto/sha256"
    "io"
)

func fileChecksum(file os.File) ([]byte, error) {
    hash := sha256.New()

    if _, err := io.Copy(hash, file); err != nil {
        return nil, err
    }

    return hash.Sum(nil), nil;
}

func stringChecksum(data string) []byte {
    hash := sha256.New()

    hash.Write([]byte(data))

    return hash.Sum(nil);
}
