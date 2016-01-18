package storage

import (
	"crypto/tls"
	"github.com/spf13/viper"
	"github.com/transhift/appdir"
	"github.com/transhift/transhift/common/security"
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
	if err = prepareConfig(dir); err != nil {
		return
	}
	host = viper.GetString("puncher.host")
	port = viper.GetInt("puncher.port")
	return
}
