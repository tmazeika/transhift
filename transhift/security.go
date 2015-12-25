package transhift

import (
    "crypto/sha256"
)

func calculateStringHash(x string) []byte {
    hash := sha256.New()
    hash.Write([]byte(x))
    return hash.Sum(nil)
}
