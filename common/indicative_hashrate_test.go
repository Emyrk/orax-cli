package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIndicativeHashRate(t *testing.T) {
	require := require.New(t)

	SaveIndicativeHashRate(1, 10000, time.Duration(1)*time.Second)
	SaveIndicativeHashRate(2, 22000, time.Duration(1)*time.Second)
	SaveIndicativeHashRate(3, 2100000, time.Duration(60)*time.Second)

	require.Equal(GetIndicativeHashRate(1), int64(10000))
	require.Equal(GetIndicativeHashRate(2), int64(22000))
	require.Equal(GetIndicativeHashRate(3), int64(35000))
	require.Equal(GetIndicativeHashRate(4), int64(0))

}
