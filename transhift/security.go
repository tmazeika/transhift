package transhift

import (
    "crypto/sha256"
    "os"
    "io"
    "crypto/rsa"
    "crypto/rand"
    "crypto/x509"
    "time"
    "math/big"
    "encoding/pem"
)

func calculateFileChecksum(file *os.File) ([]byte, error) {
    hash := sha256.New()

    if _, err := io.Copy(hash, file); err != nil {
        return nil, err
    }

    return hash.Sum(nil), nil
}

func calculateBytesHash(bytes []byte) ([]byte, error) {
    hash := sha256.New()

    if _, err := hash.Write(bytes); err != nil {
        return nil, err
    }

    return hash.Sum(nil), nil
}

func createCertificate() (keyData []byte, certData []byte, err error) {
    const RSABits = 4096
    key, err := rsa.GenerateKey(rand.Reader, RSABits)

    if err != nil {
        return nil, nil, err
    }

    keyData = x509.MarshalPKCS1PrivateKey(key)
    ca := &x509.Certificate{
        SerialNumber:          big.NewInt(50977),
        SignatureAlgorithm:    x509.SHA512WithRSA,
        PublicKeyAlgorithm:    x509.RSA,
        NotBefore:             time.Now(),
        NotAfter:              time.Now().AddDate(1000, 0, 0),
        BasicConstraintsValid: true,
        IsCA:                  true,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
        KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
    }
    certData, err = x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)

    if err != nil {
        return nil, nil, err
    }

    keyData = pem.EncodeToMemory(&pem.Block{
        Type: "RSA PRIVATE KEY",
        Bytes: keyData,
    })

    certData = pem.EncodeToMemory(&pem.Block{
        Type: "CERTIFICATE",
        Bytes: certData,
    })

    return
}
