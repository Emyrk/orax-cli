package cmd

import (
	"crypto/rand"
	"errors"
	"net/mail"
	"os"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/pbernier3/orax-cli/api"
	"gitlab.com/pbernier3/orax-cli/common"
	ed "golang.org/x/crypto/ed25519"
)

func init() {
	rootCmd.AddCommand(registerCmd)
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register miner",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.ReadInConfig()
		if err == nil && viper.IsSet("id") {
			log.Warnf("A miner identity is already configured in [%s]. Aborting registration.", common.ConfigFilePath)
		} else {
			register()
		}
	},
}

func register() {
	pub, sec, err := ed.GenerateKey(rand.Reader)

	if err != nil {
		log.Fatalf("Failed to generate a key pair: %s", err)
	}

	viper.Set("public_key", string(pub))
	viper.Set("private_key", string(sec))

	// Ask for payout address
	address, err := askPayoutAddress()
	if err != nil {
		log.Error(err)
		return
	}
	// Ask for email
	email, err := askEmail()
	if err != nil {
		log.Error(err)
		return
	}

	user, err := api.RegisterUser(email, pub, address)
	if err != nil {
		log.Errorf("Failed to register: %s", err)
		os.Exit(1)
	}

	viper.Set("id", user.ID)
	viper.Set("miner_secret", user.MinerSecret)

	err = viper.WriteConfig()
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("Registration completed. Config stored in [%s]", common.ConfigFilePath)
}

func askPayoutAddress() (address string, err error) {
	prompt := promptui.Prompt{
		Label: "Address to pay rewards to",
		// Validate: validate,
	}

	address, err = prompt.Run()
	return address, err
}

func askEmail() (email string, err error) {
	prompt := promptui.Prompt{
		Label: "Email address (for contact and recovery)",
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
