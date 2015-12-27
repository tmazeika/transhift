package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
    "crypto/rsa"
    "crypto/rand"
    "crypto/x509"
    "encoding/pem"
    "io/ioutil"
)

type Storage struct {
    customDir string

    config Config
}

type Config struct {
    PuncherHost string
    PuncherPort uint16
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

    return getFile(filepath.Join(dir, FileName))
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

    return config, nil
}

func (s *Storage) Crypto() error {
    const KeyBits = 4096
    const PrivFileName = "priv.pem"
    const PubFileName = "pub.pem"
    dir, err := s.Dir()

    if err != nil {
        return nil, err
    }

    privFilePath := filepath.Join(dir, PrivFileName)
    pubFilePath := filepath.Join(dir, PubFileName)

    var priv *rsa.PrivateKey

    if ! fileExists(privFilePath, false) {
        priv, err = rsa.GenerateKey(rand.Reader, KeyBits)

        if err != nil {
            return err
        }

        privPemData := pem.EncodeToMemory(&pem.Block{
            Type: "RSA PRIVATE KEY",
            Bytes: x509.MarshalPKCS1PrivateKey(priv),
        })

        err = ioutil.WriteFile(privFilePath, privPemData, 0600)

        if err != nil {
            return err
        }
    }

    if ! fileExists(pubFilePath, false) {
        pub, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)

        if err != nil {
            return err
        }

        pubPemData := pem.EncodeToMemory(&pem.Block{
            Type: "RSA PUBLIC KEY",
            Bytes: pub,
        })

        err = ioutil.WriteFile(pubFilePath, pubPemData, 0644)

        if err != nil {
            return err
        }
    }

    return nil
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
