package api

import (
	"errors"
	"math/big"
	"time"

	"github.com/dustin/go-humanize"
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
	Height        int64       `json:"height"`
	NbMiners      int         `json:"nbMiners"`
	NbUsers       int         `json:"nbUsers"`
	TotalHashRate BigIntStr   `json:"totalHashrate"`
	Ranks         []int       `json:"ranks"`
	Reward        int64       `json:"reward"`
	OraxReward    int64       `json:"oraxReward"`
	UsersReward   int64       `json:"usersReward"`
	UserDetail    *UserDetail `json:"userDetail"`
}

type UserDetail struct {
	HashRate BigIntStr `json:"hashrate"`
	Share    float64   `json:"share"`
	Reward   float64   `json:"reward"`
}

type BigIntStr struct {
	I string `json:"i"`
}

func (bis BigIntStr) ToString() string {
	// Try to humaize string if not too big
	bi := new(big.Int)
	bi.UnmarshalText([]byte(bis.I))
	if bi.IsInt64() {
		return humanize.Comma(bi.Int64())
	}
	return bis.I
}
