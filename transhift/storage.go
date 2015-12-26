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
    configFile, err := ConfigFile()

    if err != nil {
        return nil, err
    }

    defer configFile.Close()

    config := &Config{}
    err = json.NewDecoder(configFile).Decode(config)

    if err != nil {
        return nil, err
    }

    return config, nil
}
