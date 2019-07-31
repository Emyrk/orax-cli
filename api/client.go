package api

import (
	"fmt"
	"net/url"
	"os"

	"gitlab.com/pbernier3/orax-cli/common"
	"gopkg.in/resty.v1"
)

var OraxApiBaseUrl = "http://localhost:2666"

var log = common.GetLog()

type User struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	PayoutAddress string `json:"payoutAddress"`
}
type Miner struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type Error struct {
	Message string `json:"error"`
}

func init() {
	if os.Getenv("ORAX_API_ENDPOINT") != "" {
		_, err := url.ParseRequestURI(os.Getenv("ORAX_API_ENDPOINT"))
		if err != nil {
			log.Fatalf("Failed to parse ORAX_API_ENDPOINT: %s", err)
		}
	}
}

func RegisterUser(email string, password string, payoutAddress string) (*User, error) {
	log.Info("Registering new Orax user...")

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"email":         email,
			"payoutAddress": payoutAddress,
			"password":      password,
		}).
		SetError(&Error{}).
		SetResult(&User{}).
		Post(OraxApiBaseUrl + "/user")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		errorMsg := resp.Error().(*Error).Message
		return nil, fmt.Errorf("%s: %s", resp.Status(), errorMsg)
	}

	return resp.Result().(*User), nil
}

func GetUser(id string, password string) (*User, error) {
	resp, err := resty.R().
		SetBasicAuth(id, password).
		SetError(&Error{}).
		SetResult(&User{}).
		Get(OraxApiBaseUrl + "/user/" + id)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		errorMsg := resp.Error().(*Error).Message
		return nil, fmt.Errorf("%s: %s", resp.Status(), errorMsg)
	}

	return resp.Result().(*User), nil
}

func RegisterMiner(userId string, password string, alias string) (*Miner, error) {
	log.Info("Registering new miner with Orax...")

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"userId":   userId,
			"alias":    alias,
			"password": password,
		}).
		SetError(&Error{}).
		SetResult(&Miner{}).
		Post(OraxApiBaseUrl + "/miner")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		errorMsg := resp.Error().(*Error).Message
		return nil, fmt.Errorf("%s: %s", resp.Status(), errorMsg)
	}

	return resp.Result().(*Miner), nil
}
