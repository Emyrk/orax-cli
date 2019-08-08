package api

import (
	"errors"
	"time"
)

var ErrAuth = errors.New("Failed required authentication")

type RegisterUserResult struct {
	ID  string `json:"id"`
	JWT string `json:"jwt"`
}

type AuthenticateResult struct {
	ID  string `json:"id"`
	JWT string `json:"jwt"`
}

type RegisterMinerResult struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

type ApiError struct {
	Message string `json:"error"`
	Code    int    `json:"code"`
}

type UserInfoResult struct {
	User   User        `json:"user"`
	Miners []Miner     `json:"miners"`
	Stats  []BlockStat `json:"stats"`
}

type User struct {
	RegistrationDate time.Time `json:"registrationDate"`
	Email            string    `json:"email"`
	PayoutAddress    string    `json:"payoutAddress"`
	Balance          float64   `json:"balance"`
}

type Miner struct {
	RegistrationDate     time.Time `json:"registrationDate"`
	Alias                string    `json:"alias"`
	LatestSubmissionDate time.Time `json:"latestSubmissionDate"`
	LatestOpCount        int64     `json:"latestOpCount"`
	LatestDuration       int64     `json:"latestDuration"`
}

type BlockStat struct {
	Height       int64       `json:"height"`
	MinerCount   int         `json:"minerCount"`
	Duration     int64       `json:"duration"`
	TotalOpCount int64       `json:"totalOpCount"`
	Ranks        []int       `json:"ranks"`
	Reward       int64       `json:"reward"`
	OraxReward   int64       `json:"oraxReward"`
	UsersReward  int64       `json:"usersReward"`
	UserDetail   *UserDetail `json:"userDetail"`
}

type UserDetail struct {
	OpCount  int64   `json:"opCount"`
	Duration int64   `json:"duration"`
	Share    float64 `json:"share"`
	Reward   float64 `json:"reward"`
}
