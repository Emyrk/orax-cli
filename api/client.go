package api

import (
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/spf13/viper"
	"gitlab.com/pbernier3/orax-cli/common"
	"gopkg.in/resty.v1"
)

var oraxAPIBaseURL string

var log = common.GetLog()

func init() {
	// Override endpoint with env variable
	if os.Getenv("ORAX_API_ENDPOINT") != "" {
		_, err := url.ParseRequestURI(os.Getenv("ORAX_API_ENDPOINT"))
		if err != nil {
			log.Fatalf("Failed to parse ORAX_API_ENDPOINT: %s", err)
		}
		oraxAPIBaseURL = os.Getenv("ORAX_API_ENDPOINT")
	} else if oraxAPIBaseURL == "" {
		// If not set at build time fallback to local dev endpoint
		oraxAPIBaseURL = "http://localhost:2666"
	}
}

func RegisterUser(email string, password string, payoutAddress string) (*RegisterUserResult, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]string{
			"email":         email,
			"payoutAddress": payoutAddress,
			"password":      password,
		}).
		SetError(&ApiError{}).
		SetResult(&RegisterUserResult{}).
		Post(oraxAPIBaseURL + "/user")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		errorMsg := resp.Error().(*ApiError).Message
		return nil, fmt.Errorf("%s: %s", resp.Status(), errorMsg)
	}

	return resp.Result().(*RegisterUserResult), nil
}

// Authenticate user and returns a JSON Web Token.
// Input `id` can either be user id or email.
func Authenticate(id string, password string) (*AuthenticateResult, error) {
	resp, err := resty.R().
		SetBasicAuth(id, password).
		SetError(&ApiError{}).
		SetResult(&AuthenticateResult{}).
		Post(oraxAPIBaseURL + "/user/auth")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		errorMsg := resp.Error().(*ApiError).Message
		return nil, fmt.Errorf("%s: %s", resp.Status(), errorMsg)
	}

	return resp.Result().(*AuthenticateResult), nil
}

func RegisterMiner(alias string) (*RegisterMinerResult, error) {
	resp, err := resty.R().
		SetHeader("Content-Type", "application/json").
		SetAuthToken(viper.GetString("jwt")).
		SetBody(map[string]string{
			"alias": alias,
		}).
		SetError(&ApiError{}).
		SetResult(&RegisterMinerResult{}).
		Post(oraxAPIBaseURL + "/miner")

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		errorMsg := resp.Error().(*ApiError).Message
		return nil, fmt.Errorf("%s: %s", resp.Status(), errorMsg)
	}

	return resp.Result().(*RegisterMinerResult), nil
}

func GetUserInfo(id string, height int, pageSize int) (*UserInfoResult, error) {
	resp, err := resty.R().
		SetAuthToken(viper.GetString("jwt")).
		SetQueryParam("height", strconv.Itoa(height)).
		SetQueryParam("pageSize", strconv.Itoa(pageSize)).
		SetError(&ApiError{}).
		SetResult(&UserInfoResult{}).
		Get(oraxAPIBaseURL + "/user/" + id)

	if err != nil {
		return nil, err
	}

	if resp.IsError() {
		apiError := resp.Error().(*ApiError)
		if apiError.Code == 1 {
			return nil, ErrAuth
		}
		return nil, fmt.Errorf("%s: %s", resp.Status(), apiError.Message)
	}

	return resp.Result().(*UserInfoResult), nil
}
