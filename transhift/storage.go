package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
    "io/ioutil"
    "fmt"
    "crypto/x509"
    "crypto/tls"
)

type Config struct {
    PuncherHost string
    PuncherPort uint16
}

func (c Config) PuncherPortStr() string {
    return fmt.Sprint(c.PuncherPort)
}

type Storage struct {
    customDir string

    config    *Config
    cert      x509.Certificate
}

func (s Storage) Dir() (string, error) {
    const DefDirName = ".transhift"

    if len(s.customDir) == 0 {
        user, err := user.Current()

        if err != nil {
            return "", err
        }

        return getDir(filepath.Join(user.HomeDir, DefDirName))
    } else {
        return getDir(s.customDir)
    }
}

func (s Storage) ConfigFile() (*os.File, error) {
    const FileName = "config.json"
    dir, err := s.Dir()

    if err != nil {
        return nil, err
    }

    filePath := filepath.Join(dir, FileName)

    if ! fileExists(filePath, false) {
        defConfig := Config{
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
    if s.config != nil {
        return s.config, nil
    }

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

func (s Storage) Certificate() (tls.Certificate, error) {
    const CertFileName = "cert.pem"
    const KeyFileName = "cert.key"
    dir, err := s.Dir()

    if err != nil {
        return tls.Certificate{}, err
    }

    certFilePath := filepath.Join(dir, CertFileName)
    keyFilePath := filepath.Join(dir, KeyFileName)

    if ! fileExists(certFilePath, false) || ! fileExists(keyFilePath, false) {
        fmt.Print("Generating crypto... ")

        keyData, certData, err := createCertificate()

        if err != nil {
            return tls.Certificate{}, err
        }

        err = ioutil.WriteFile(certFilePath, certData, 0600)

        if err != nil {
            return tls.Certificate{}, err
        }

        err = ioutil.WriteFile(keyFilePath, keyData, 0600)

        if err != nil {
            return tls.Certificate{}, err
        }

        fmt.Println("done")
    }

    return tls.LoadX509KeyPair(certFilePath, keyFilePath)
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
