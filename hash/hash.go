package hash

import lxr "github.com/pegnet/LXR256"

var LX lxr.LXRHash

func init() {
	LX.Init(0xfafaececfafaecec, 25, 256, 5)
}

func Hash(data []byte) []byte {
	return LX.Hash(data)
}
