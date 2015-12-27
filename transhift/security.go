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
    "crypto/tls"
)

func calculateFileChecksum(file *os.File) []byte {
    hash := sha256.New()
    io.Copy(hash, file)
    return hash.Sum(nil)
}

func createCertificate() (key *rsa.PrivateKey, keyData []byte, certData []byte, err error) {
    const RSABits = 4096
    key, err = rsa.GenerateKey(rand.Reader, RSABits)

    if err != nil {
        return nil, nil, nil, err
    }

    keyData = x509.MarshalPKCS1PrivateKey(key)
    ca := &x509.Certificate{
        SerialNumber:          big.NewInt(50977),
        SignatureAlgorithm:    x509.SHA512WithRSA,
        PublicKeyAlgorithm:    x509.RSA,
        NotBefore:             time.Now(),
        NotAfter:              time.Now().AddDate(1, 0, 0),
        BasicConstraintsValid: true,
        IsCA:                  true,
        ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
        KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
    }
    certData, err = x509.CreateCertificate(rand.Reader, ca, ca, &key.PublicKey, key)

    if err != nil {
        return nil, nil, nil, err
    }

    return
}

func parseCertificate(keyData, certData []byte) (key *rsa.PrivateKey, cert *x509.Certificate, certPool *x509.CertPool, err error) {
    key, err = x509.ParsePKCS1PrivateKey(keyData)

    if err != nil {
        return nil, nil, nil, err
    }

    cert, err = x509.ParseCertificate(certData)

    if err != nil {
        return nil, nil, nil, err
    }

    certPool = x509.NewCertPool()
    err = nil

    certPool.AddCert(cert)
    return
}

func createTLSConfig(key *rsa.PrivateKey, certPool *x509.CertPool) *tls.Config {
    return &tls.Config{
        Certificates: []tls.Certificate{tls.Certificate{
            Certificate: certPool.Subjects(),
            PrivateKey: key,
        }},
        MinVersion: tls.VersionTLS12,
    }
}
