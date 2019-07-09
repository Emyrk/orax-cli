package common

import (
	"log"
	"os"
	"os/user"

	"github.com/spf13/viper"
)

var ConfigFilePath string

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	configFolderPath := usr.HomeDir + "/.orax"
	ConfigFilePath = configFolderPath + "/config.yml"
	if _, err := os.Stat(ConfigFilePath); os.IsNotExist(err) {
		os.MkdirAll(configFolderPath, 0700)
		os.Create(ConfigFilePath)
	}

	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	viper.AddConfigPath(configFolderPath)
}
