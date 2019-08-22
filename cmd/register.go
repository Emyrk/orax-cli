package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/fatih/color"
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
			color.Red("A miner identity is already configured in [%s]. Aborting registration.", configFilePath)
		} else {
			// Write a blank config file early to verify it's possible
			// (permission, file extension...)
			err = viper.SafeWriteConfig()
			if err != nil {
				color.Red(err.Error())
				os.Exit(1)
			}
			err := register()
			if err != nil {
				fmt.Printf("\n")
				color.Red(err.Error())
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

	color.Green("\nRegistration completed. Config stored in [%s]\n\n", configFilePath)

	return nil
}

func getAccountChoice() (string, error) {
	if runtime.GOOS == "windows" {
		prompt := promptui.Prompt{
			Label: "Use (n)ew account or (e)xisting account? [n/e]",
			Validate: func(input string) error {
				choice := strings.ToLower(input)
				if choice == "e" || choice == "n" || choice == "new" || choice == "existing" {
					return nil
				}
				return errors.New("Invalid choice")
			},
		}

		choice, err := prompt.Run()
		if err != nil {
			return "", err
		}

		if choice == "n" {
			return "new", nil
		} else if choice == "e" {
			return "existing", nil
		}
		return choice, nil
	}

	prompt := promptui.Select{
		Label: "Register miner with",
		Items: []string{"a new Orax account", "an existing Orax account"},
	}

	i, _, err := prompt.Run()

	if err != nil {
		return "", err
	}

	if i == 0 {
		return "new", nil
	}
	return "existing", nil
}

func getOraxUser() (err error) {
	choice, err := getAccountChoice()
	if err != nil {
		return err
	}

	var userID, jwt string
	fmt.Printf("\n")
	if choice == "new" {
		userID, jwt, err = newOraxUser()
		if err != nil {
			return err
		}
		color.Green("\nNew Orax user registered successfully.\n\n")
	} else {
		userID, jwt, err = existingOraxUser()

		if err != nil {
			return err
		}
		color.Green("\nSuccessfully authenticated.\n\n")
	}

	viper.Set("user_id", userID)
	viper.Set("jwt", jwt)

	return nil
}

func registerMiner() error {
	fmt.Println("Registering this machine as a miner linked to your account:")

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
