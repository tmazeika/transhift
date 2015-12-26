package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
)

type Config struct {
    PuncherHost string
    PuncherPort uint16
}

func ConfigFile() (*os.File, error) {
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

func (c *Config) Write() error {
    file, err := ConfigFile()

    file.Close()
    os.Remove(file.Name())

    if err != nil {
        return nil
    }

    file, err = ConfigFile()
    defer file.Close()

    if err != nil {
        return err
    }

    return json.NewEncoder(file).Encode(c)
}
