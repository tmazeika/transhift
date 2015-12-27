package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
    "crypto/rsa"
    "io/ioutil"
    "fmt"
    "crypto/x509"
)

type Config struct {
    PuncherHost string
    PuncherPort uint16
}

func (c *Config) PuncherPortStr() string {
    return fmt.Sprint(c.PuncherPort)
}

type Storage struct {
    customDir string

    config    *Config
    key       *rsa.PrivateKey
    cert      *x509.Certificate
    certPool  *x509.CertPool
}

func (s *Storage) Dir() (string, error) {
    const DefDirName = ".transhift"

    if s.customDir == "" {
        user, err := user.Current()

        if err != nil {
            return "", err
        }

        return getDir(filepath.Join(user.HomeDir, DefDirName))
    } else {
        return getDir(s.customDir)
    }
}

func (s *Storage) ConfigFile() (*os.File, error) {
    const FileName = "config.json"
    dir, err := s.Dir()

    if err != nil {
        return nil, err
    }

    filePath := filepath.Join(dir, FileName)

    if ! fileExists(filePath, false) {
        defConfig := &Config{
            PuncherHost: "127.0.0.1", // TODO: change
            PuncherPort: 50977,
        }

        data, err := json.MarshalIndent(defConfig, "", "  ")

        if err != nil {
            return nil, err
        }

        err = ioutil.WriteFile(filePath, data, 0644)

        if err != nil {
            return nil, err
        }
    }

    return getFile(filePath)
}

func (s *Storage) Config() (*Config, error) {
    file, err := s.ConfigFile()

    if err != nil {
        return nil, err
    }

    defer file.Close()
    config := &Config{}
    err = json.NewDecoder(file).Decode(config)

    if err != nil {
        return nil, err
    }

    s.config = config
    return config, nil
}

func (s *Storage) Crypto() (*rsa.PrivateKey, *x509.Certificate, *x509.CertPool, error) {
    const KeyFileName = "key"
    const CertFileName = "certificate"
    dir, err := s.Dir()

    if err != nil {
        return nil, nil, nil, err
    }

    keyFilePath := filepath.Join(dir, KeyFileName)
    certFilePath := filepath.Join(dir, CertFileName)

    if ! fileExists(keyFilePath, false) || ! fileExists(certFilePath, false) {
        fmt.Print("Generating crypto... ")

        var keyData, certData []byte
        s.key, keyData, certData, err = createCertificate()

        if err != nil {
            return nil, nil, nil, err
        }

        err = ioutil.WriteFile(keyFilePath, keyData, 0600)

        if err != nil {
            return nil, nil, nil, err
        }

        err = ioutil.WriteFile(certFilePath, certData, 0600)

        if err != nil {
            return nil, nil, nil, err
        }

        fmt.Println("done")
    }

    keyFile, err := getFile(keyFilePath)

    if err != nil {
        return nil, nil, nil, err
    }

    defer keyFile.Close()
    certFile, err := getFile(certFilePath)

    if err != nil {
        return nil, nil, nil, err
    }

    defer certFile.Close()
    keyData, err := ioutil.ReadAll(keyFile)

    if err != nil {
        return nil, nil, nil, err
    }

    certData, err := ioutil.ReadAll(certFile)

    if err != nil {
        return nil, nil, nil, err
    }

    s.key, s.cert, s.certPool, err = parseCertificate(keyData, certData)

    if err != nil {
        return nil, nil, nil, err
    }

    return s.key, s.cert, s.certPool, nil
}

func getFile(path string) (*os.File, error) {
    if fileExists(path, false) {
        return os.Open(path)
    }

    return os.Create(path)
}

func getDir(path string) (string, error) {
    if fileExists(path, true) {
        return path, nil
    }

    err := os.MkdirAll(path, 0700)

    if err != nil {
        return "", err
    }

    return path, nil
}

func fileExists(path string, asDir bool) bool {
    info, err := os.Stat(path)

    if err != nil {
        return false
    }

    if asDir {
        return info.IsDir()
    }

    return info.Mode().IsRegular()
}
