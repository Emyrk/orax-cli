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

var usernameFlag, passwordFlag, aliasFlag string

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVarP(&usernameFlag, "username", "u", "", "Orax account username (email).")
	registerCmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Orax account password.")
	registerCmd.Flags().StringVarP(&aliasFlag, "alias", "a", "", "Miner alias.")
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
			err = viper.WriteConfig()
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
	if usernameFlag != "" && passwordFlag != "" && aliasFlag != "" {
		return registerNonInteractive()
	}

	return registerInteractive()

}

func registerNonInteractive() error {
	result, err := api.Authenticate(usernameFlag, passwordFlag)
	if err != nil {
		return fmt.Errorf("Failed to authenticate: %s", err)
	}
	color.Green("\nSuccessfully authenticated.")

	viper.Set("user_id", result.ID)
	viper.Set("jwt", result.JWT)

	err = registerMiner(aliasFlag)
	if err != nil {
		return err
	}

	return saveConfiguration()
}

func registerInteractive() error {
	err := getOraxUser()
	if err != nil {
		return err
	}

	err = registerMinerPrompt()
	if err != nil {
		return err
	}

	return saveConfiguration()
}

func saveConfiguration() error {
	err := viper.WriteConfig()
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

func registerMinerPrompt() error {
	fmt.Println("Registering this machine as a miner linked to your account:")

	alias, err := askMinerAlias()
	if err != nil {
		return err
	}

	return registerMiner(alias)
}

func registerMiner(alias string) error {
	miner, err := api.RegisterMiner(alias)
	if err != nil {
		return err
	}

	viper.Set("miner_id", miner.ID)
	viper.Set("miner_secret", miner.Secret)

	return nil
}
