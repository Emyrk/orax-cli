package hash

import (
	"sync"

	lxr "github.com/pegnet/LXR256"
)

var LX lxr.LXRHash
var once sync.Once

func InitLXR() {
	once.Do(func() {
		LX.Init(0xfafaececfafaecec, 25, 256, 5)
	})
}

func Hash(data []byte) []byte {
	return LX.Hash(data)
}
