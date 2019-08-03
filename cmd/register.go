package cmd

import (
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/pbernier3/orax-cli/api"
)

func init() {
	rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register miner",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.ReadInConfig()
		if err == nil && viper.IsSet("miner_id") {
			log.Warnf("A miner identity is already configured in [%s]. Aborting registration.", configFilePath)
		} else {
			err := register()
			if err != nil {
				log.Error(err)
				os.Exit(1)
			}
		}
	},
}

func register() error {
	err := getOraxUser()
	if err != nil {
		return err
	}

	err = registerMiner()
	if err != nil {
		return err
	}

	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	log.Infof("Registration completed. Config stored in [%s]", configFilePath)

	return nil
}

func getOraxUser() (err error) {
	prompt := promptui.Select{
		Label: "Register miner with",
		Items: []string{"a new Orax account", "an existing Orax account"},
	}

	i, _, err := prompt.Run()

	if err != nil {
		return err
	}

	var userID, jwt string
	if i == 0 {
		userID, jwt, err = newOraxUser()
	} else {
		userID, jwt, err = existingOraxUser()
	}

	if err != nil {
		return err
	}

	viper.Set("user_id", userID)
	viper.Set("jwt", jwt)

	return nil
}

func registerMiner() error {
	log.Info("Registering this machine as a miner linked towith your account:")

	alias, err := askMinerAlias()
	if err != nil {
		return err
	}

	miner, err := api.RegisterMiner(alias)
	if err != nil {
		return err
	}

	viper.Set("miner_id", miner.ID)
	viper.Set("miner_secret", miner.Secret)

	return nil
}
