package api

import (
	"encoding/base64"
	"fmt"

	"gitlab.com/pbernier3/orax-cli/common"
	"gopkg.in/resty.v1"
)

const OraxApiBaseUrl = "http://localhost:2666"

var log = common.GetLog()

type User struct {
	Id          string `json:"id"`
	MinerSecret string `json:"minerSecret"`
}

type Error struct {
	Message string `json:"error"`
}

func RegisterUser(email string, publicKey []byte, payoutAddress string) (*User, error) {
	log.Info("Registering miner with Orax server...")

	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"email":         email,
			"payoutAddress": payoutAddress,
			"publicKey":     base64.StdEncoding.EncodeToString(publicKey),
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
