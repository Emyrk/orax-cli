package cmd

import (
	"fmt"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/pbernier3/orax-cli/common"
)

var rootCmd = &cobra.Command{
	Use:   "orax",
	Short: "Mining client for the Orax mining pool",
}

var configFilePath string

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Config file path (default $HOME/.orax/config.yml)")
}

var log = common.GetLog()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initConfig() {

	if configFilePath != "" {
		viper.SetConfigFile(configFilePath)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		configFolderPath := home + "/.orax"
		configFilePath = configFolderPath + "/config.yml"
		// Create .orax folder and empty config file if necessary
		if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
			os.MkdirAll(configFolderPath, 0700)
			os.Create(configFilePath)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath(configFolderPath)
	}
}
