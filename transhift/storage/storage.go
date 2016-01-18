package storage

import (
	"crypto/tls"
	"github.com/spf13/viper"
	"github.com/transhift/appdir"
	"github.com/transhift/transhift/common/security"
	"os"
)

func Prepare(appDirPath string) (host string, port int, cert tls.Certificate, err error) {
	const DefAppDir = "$HOME/.transhift"
	dir, err := appdir.NewPreferNonEmpty(appDirPath, DefAppDir)
	if err != nil {
		return
	}

	const KeyName = "cert.key"
	const CertName = "cert.pem"
	if cert, err = security.Certificate(KeyName, CertName, dir); err != nil {
		return
	}

	const ConfigName = "config.json"
	file, err := dir.IfNExistsThenGet(ConfigName, func(file *os.File) (b []byte, err error) {
		s :=
// Default config.
`{
  "puncher": {
    "host": "104.236.76.95",
    "port": 50977
  }
}`
		return []byte(s), nil
	})
	if err != nil {
		return
	}
	defer file.Close()

	viper.SetConfigFile(file.Name())
	if err = viper.ReadInConfig(); err != nil {
		return
	}
	host = viper.GetString("puncher.host")
	port = viper.GetInt("puncher.port")
	return
}
