package common

import (
	"strconv"
	"time"

	"github.com/spf13/viper"
)

func GetIndicativeHashRate(nbMiners int) int64 {
	return viper.GetInt64(buildKey(nbMiners))
}

func SaveIndicativeHashRate(nbMiners int, totalOps int64, duration time.Duration) error {
	hashRate := int64(float64(totalOps) / duration.Seconds())
	key := buildKey(nbMiners)
	viper.Set(key, hashRate)
	return viper.WriteConfig()
}

func buildKey(nbMiners int) string {
	return "hash_rate_" + strconv.Itoa(nbMiners)
}
