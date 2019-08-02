package cmd

import (
	"errors"
	"fmt"
	"net/mail"
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

	viper.Set("user_id", userID)
	viper.Set("jwt", jwt)

	return err
}

func existingOraxUser() (userID string, jwt string, err error) {
	email, err := askEmail()
	if err != nil {
		return "", "", err
	}
	password, err := askPassword()
	if err != nil {
		return "", "", err
	}

	result, err := api.Authenticate(email, password)
	if err != nil {
		return "", "", fmt.Errorf("Failed to authenticate: %s", err)
	}

	return result.ID, result.JWT, nil
}

func newOraxUser() (userID string, jwt string, err error) {
	email, err := askEmail()
	if err != nil {
		return "", "", err
	}
	password, err := askPassword()
	if err != nil {
		return "", "", err
	}
	payoutAddress, err := askPayoutAddress()
	if err != nil {
		return "", "", err
	}

	user, err := api.RegisterUser(email, password, payoutAddress)
	if err != nil {
		return "", "", fmt.Errorf("Failed to register a new Orax user: %s", err)
	}

	log.Info("New Orax user registered.")

	return user.ID, user.JWT, nil
}

func registerMiner() error {
	log.Info("Registering this machine as a miner with your account:")

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

/****************
* ask functions
****************/

func askEmail() (email string, err error) {
	prompt := promptui.Prompt{
		Label: "Email address",
		Validate: func(input string) error {
			_, err := mail.ParseAddress(input)
			if err != nil {
				return errors.New("Invalid email address")
			}
			return nil
		},
	}

	email, err = prompt.Run()
	return email, err
}

func askPassword() (password string, err error) {
	prompt := promptui.Prompt{
		Label: "Password",
		Mask:  '*',
		Validate: func(input string) error {
			if len(input) < 8 {
				return errors.New("Password must have more than 8 characters")
			}
			return nil
		},
	}

	password, err = prompt.Run()
	return password, err
}

func askPayoutAddress() (address string, err error) {
	prompt := promptui.Prompt{
		Label: "Address to pay rewards to",
		// TODO Validate: validate,
	}

	address, err = prompt.Run()
	return address, err
}

func askMinerAlias() (alias string, err error) {
	prompt := promptui.Prompt{
		Label: "Miner alias (name of the machine for instance)",
	}

	alias, err = prompt.Run()
	return alias, err
}
