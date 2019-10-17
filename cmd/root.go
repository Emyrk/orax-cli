package cmd

import (
	"log"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/oraxpool/orax-cli/common"
)

var rootCmd = &cobra.Command{
	Use:   "orax-cli",
	Short: "Mining client for the Orax mining pool",
}

var (
	configFilePath string
	logColor       string
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = common.Version
	rootCmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", "", "Config file path (default $HOME/.orax/config.yml)")
	rootCmd.PersistentFlags().StringVar(&logColor, "color", "auto", "Log color: [auto|on|off]")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		common.PrintError(err.Error())
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
			os.OpenFile(configFilePath, os.O_CREATE, 0600)
		}

		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
		viper.AddConfigPath(configFolderPath)
	}

	common.SetLogColor(logColor)
}
