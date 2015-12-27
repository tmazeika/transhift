package transhift

import (
    "crypto/sha256"
    "os"
    "io"
    "crypto/rsa"
    "crypto/rand"
    "encoding/pem"
    "crypto/x509"
)

func calculateFileChecksum(file *os.File) []byte {
    hash := sha256.New()
    io.Copy(hash, file)
    return hash.Sum(nil)
}

func generatePrivateKey() (privKey *rsa.PrivateKey, pemData []byte, err error) {
    const BitCount = 4096
    privKey, err = rsa.GenerateKey(rand.Reader, BitCount)

    if err != nil {
        return nil, nil, err
    }

    pemData = pem.EncodeToMemory(&pem.Block{
        Type: "RSA PRIVATE KEY",
        Bytes: x509.MarshalPKCS1PrivateKey(privKey),
    })
    err = nil
    return
}

func getPublicKeyPem(privKey *rsa.PrivateKey) (pemData []byte, err error) {
    pubKey, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)

    if err != nil {
        return nil, err
    }

    return pem.EncodeToMemory(&pem.Block{
        Type: "RSA PUBLIC KEY",
        Bytes: pubKey,
    }), nil
}

func parsePrivateKeyPem(pemData []byte) (*rsa.PrivateKey, error) {
    block, _ := pem.Decode(pemData)
    return x509.ParsePKCS1PrivateKey(block.Bytes)
}
