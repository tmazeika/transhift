package tstorage

import (
	"github.com/spf13/viper"
	"github.com/transhift/appdir"
)

func PrepareConfig(a *appdir.AppDir) error {
	const ConfigName = "config.json"
	const DefHost = "104.236.76.95"
	const DefPort = 50977

	viper.SetConfigFile(a.FilePath(ConfigName))
	viper.SetDefault("puncher.host", DefHost)
	viper.SetDefault("puncher.port", DefPort)
	return viper.ReadInConfig()
}
