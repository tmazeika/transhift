package transhift

import (
    "os/user"
    "path/filepath"
    "os"
    "encoding/json"
)

type Storage struct {
    customDir *string

    config Config
}

type Config struct {
    PuncherHost string
    PuncherPort uint16
}

func (s *Storage) Dir() (*string, error) {
    const DefDirName = ".transhift"
    var dirPath string

    if s.customDir == nil {
        user, err := user.Current()

        if err != nil {
            return nil, err
        }

        dirPath = filepath.Join(user.HomeDir, DefDirName)
    } else {
        dirPath = *s.customDir
    }

    if info, err := os.Stat(dirPath); os.IsNotExist(err) || (info != nil && ! info.IsDir()) {
        err = os.MkdirAll(dirPath, 0700)

        if err != nil {
            return nil, err
        }
    }

    return &dirPath, nil
}

func (s *Storage) ConfigFile() (*os.File, error) {
    const FileName = "config.json"
    dir, err := s.Dir()

    if err != nil {
        return nil, err
    }

    filePath := filepath.Join(dir, FileName)

    if info, err := os.Stat(filePath); os.IsNotExist(err) || (info != nil && ! info.Mode().IsRegular()) {
        return os.Create(filePath)
    }

    return os.Open(filePath)
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
