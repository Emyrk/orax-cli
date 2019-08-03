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
	LatestHashRate       uint64    `json:"latestHashrate"`
}

type BlockStat struct {
	Height        uint        `json:"height"`
	NbMiners      uint        `json:"nbMiners"`
	NbUsers       uint        `json:"nbUsers"`
	TotalHashRate uint64      `json:"totalHashrate"`
	Ranks         []int       `json:"ranks"`
	Reward        uint64      `json:"reward"`
	OraxReward    uint64      `json:"oraxReward"`
	UsersReward   uint64      `json:"userReward"`
	UserDetail    *UserDetail `json:"userDetail"`
}

type UserDetail struct {
	HashRate uint64  `json:"hashrate"`
	Share    float64 `json:"share"`
	Reward   float64 `json:"reward"`
}
