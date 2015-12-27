package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
    "crypto/rsa"
    "io/ioutil"
    "fmt"
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

    config *Config
    privKey *rsa.PrivateKey
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

func (s *Storage) PrivKey() (*rsa.PrivateKey, error) {
    const KeyBits = 4096
    const PrivFileName = "priv.pem"
    const PubFileName = "pub.pem"
    dir, err := s.Dir()

    if err != nil {
        return nil, err
    }

    privFilePath := filepath.Join(dir, PrivFileName)
    pubFilePath := filepath.Join(dir, PubFileName)
    forceNewPubKeyWrite := false

    if ! fileExists(privFilePath, false) {
        fmt.Print("Generating keys... ")

        forceNewPubKeyWrite = true
        var pemData []byte
        s.privKey, pemData, err = generatePrivateKey()

        if err != nil {
            return nil, err
        }

        err = ioutil.WriteFile(privFilePath, pemData, 0600)

        if err != nil {
            return nil, err
        }

        fmt.Println("done")
    } else {
        privFile, err := getFile(privFilePath)

        if err != nil {
            return nil, err
        }

        defer privFile.Close()
        pemData, err := ioutil.ReadAll(privFile)

        if err != nil {
            return nil, err
        }

        s.privKey, err = parsePrivateKeyPem(pemData)

        if err != nil {
            return nil, err
        }
    }

    if forceNewPubKeyWrite || ! fileExists(pubFilePath, false) {
        pemData, err := getPublicKeyPem(s.privKey)

        if err != nil {
            return nil, err
        }

        err = ioutil.WriteFile(pubFilePath, pemData, 0644)

        if err != nil {
            return nil, err
        }
    }

    return s.privKey, nil
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
