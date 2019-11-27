package hash

import (
	"sync"

	lxr "github.com/pegnet/LXRHash"
	"gitlab.com/oraxpool/orax-cli/common"
)

var LX lxr.LXRHash
var once sync.Once

var (
	log = common.GetLog()
)

func InitLXR() {
	once.Do(func() {
		log.Info("Initializing LXR hash...")
		LX.Verbose(true)
		LX.Init(0xfafaececfafaecec, 30, 256, 5)
	})
}

func Hash(data []byte) []byte {
	return LX.Hash(data)
}
