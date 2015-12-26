package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
)

type Storage struct {
    config Config
}

func StorageDirectory() (string, error) {
    const DirName = ".transhift"

    user, err := user.Current()

    if err != nil {
        return "", err
    }

    dirPath := filepath.Join(user.HomeDir, DirName)

    if info, err := os.Stat(dirPath); os.IsNotExist(err) || (info != nil && ! info.IsDir()) {
        err = os.Mkdir(dirPath, 0700)

        if err != nil {
            return "", err
        }
    }

    return dirPath, nil
}

type Config struct {
    PuncherHost string
    PuncherPort uint16
}

func (Storage) ConfigFile() (*os.File, error) {
    const AppDir = ".transhift"
    const FileName = "config.json"

    user, err := user.Current()

    if err != nil {
        return nil, err
    }

    os.Mkdir(filepath.Join(user.HomeDir, AppDir), 0700)

    filePath := filepath.Join(user.HomeDir, AppDir, FileName)
    file, err := os.Open(filePath)

    if err != nil {
        return os.Create(filePath)
    }

    return file, nil
}

func ReadConfig() (*Config, error) {
    const AppDir = ".transhift"
    const FileName = "config.json"

    user, err := user.Current()

    if err != nil {
        return nil, err
    }

    os.Mkdir(filepath.Join(user.HomeDir, AppDir), 0700)

    filePath := filepath.Join(user.HomeDir, AppDir, FileName)
    file, err := os.Open(filePath)

    if err != nil {
        file, err = os.Create(filePath)
    }

    file, err := ConfigFile()
    defer file.Close()

    if err != nil {
        return nil, err
    }

    config := &Config{}
    err = json.NewDecoder(file).Decode(config)

    if err != nil {
        return nil, err
    }

    return config, nil
}
