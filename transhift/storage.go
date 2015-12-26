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

func StorageDir() (*string, error) {
    const DirName = ".transhift"

    user, err := user.Current()

    if err != nil {
        return nil, err
    }

    dirPath := filepath.Join(user.HomeDir, DirName)

    if info, err := os.Stat(dirPath); os.IsNotExist(err) || (info != nil && ! info.IsDir()) {
        err = os.Mkdir(dirPath, 0700)

        if err != nil {
            return nil, err
        }
    }

    return &dirPath, nil
}

func ConfigFile() (*os.File, error) {
    const FileName = "config.json"

    storageDir, err := StorageDir()

    if err != nil {
        return nil, err
    }

    filePath := filepath.Join(storageDir, FileName)

    if info, err := os.Stat(filePath); os.IsNotExist(err) || (info != nil && ! info.Mode().IsRegular()) {
        return os.Create(filePath)
    }

    return os.Open(filePath)
}

type Config struct {
    PuncherHost string
    PuncherPort uint16
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
