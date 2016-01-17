package tstorage

import "github.com/transhift/transhift/common/storage"

const (
    CertFileName = "cert.pem"
    KeyFileName = "cert.key"
)

type Config struct {
    Puncher map[string]string
}

func Prepare(appDir string) (storage.Storage, error) {
    defConf := Config{
        Puncher: map[string]string{
            "host": "104.236.76.95",
            "port": "50977",
        },
    }
    return storage.New(appDir, defConf)
}
