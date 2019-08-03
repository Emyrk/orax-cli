package cmd

import (
	"errors"
	"fmt"
	"net/mail"

	"github.com/manifoldco/promptui"
	"gitlab.com/pbernier3/orax-cli/api"
)

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

/////////////////////

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
