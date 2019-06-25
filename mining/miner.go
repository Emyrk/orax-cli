package mining

import (
	"crypto/rand"
	"encoding/binary"
	"sync"

	lxr "github.com/pegnet/LXR256"
)

type Miner struct {
	id         int
	hasher     *lxr.LXRHash
	stop       chan int
	opsCounter uint64
	bestNonce  *Nonce
}

func NewMiner(id int, hasher *lxr.LXRHash) *Miner {
	miner := new(Miner)
	miner.id = id
	miner.hasher = hasher
	miner.stop = make(chan int)
	miner.bestNonce = nil

	return miner
}

func (miner *Miner) Reset() {
	miner.opsCounter = 0
	miner.bestNonce = nil
}

type Nonce struct {
	Nonce      []byte
	Difficulty uint64
}

func (miner *Miner) mine(oprHash []byte, wg *sync.WaitGroup) {
	// Create a slice of sufficient capacity to avoid a new underlying array to be allocated
	// when appending nonce after the OPR
	dataToMine := make([]byte, 32, 64)
	copy(dataToMine, oprHash)

	nonce := make([]byte, 32)
	rand.Read(nonce)

mining:
	for i := 0; ; i++ {
		select {
		case <-miner.stop:
			break mining
		default:
		}

		// TODO: Better way?
		k := 0
		for j := i; j > 0; j = j >> 8 {
			nonce[k] = byte(j)
			k++
		}

		dataToHash := append(dataToMine, nonce...)
		h := LX.Hash(dataToHash)
		diff := computeDifficulty(h)
		miner.opsCounter++

		if miner.bestNonce == nil || diff > miner.bestNonce.Difficulty {
			miner.bestNonce = &Nonce{append([]byte(nil), nonce...), diff}
		}
	}
	wg.Done()
}

func computeDifficulty(h []byte) uint64 {
	return binary.BigEndian.Uint64(h[:8])
}
