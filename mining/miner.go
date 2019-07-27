package mining

import (
	"crypto/rand"
	"encoding/binary"
	"sync"

	"gitlab.com/pbernier3/orax-cli/hash"
)

type Miner struct {
	id         int
	stop       chan int
	opsCounter uint64
	bestNonces []*Nonce
}

func NewMiner(id int) *Miner {
	miner := new(Miner)
	miner.id = id
	miner.stop = make(chan int)
	miner.bestNonces = make([]*Nonce, 0, 256)

	return miner
}

func (miner *Miner) Reset() {
	miner.opsCounter = 0
	miner.bestNonces = make([]*Nonce, 0, 256)
}

type Nonce struct {
	Nonce      []byte
	Difficulty uint64
}

func (miner *Miner) mine(oprHash []byte, maxNonces int, wg *sync.WaitGroup) {
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
		h := hash.Hash(dataToHash)
		diff := computeDifficulty(h)
		miner.opsCounter++

		if len(miner.bestNonces) < maxNonces {
			// If the buffer is not yet full just append
			miner.bestNonces = append(miner.bestNonces, &Nonce{copyNonce(nonce), diff})
			SortNoncesByDiff(miner.bestNonces)
		} else if miner.bestNonces[len(miner.bestNonces)-1].Difficulty < diff {
			// Otherwise if diff is better than the last best nonce
			miner.bestNonces = miner.bestNonces[:maxNonces-1]
			miner.bestNonces = append(miner.bestNonces, &Nonce{copyNonce(nonce), diff})
			SortNoncesByDiff(miner.bestNonces)
		}
	}

	wg.Done()
}

func copyNonce(nonce []byte) []byte {
	copy := make([]byte, len(nonce))
	for i, b := range nonce {
		copy[i] = b
	}
	return copy
}

func computeDifficulty(h []byte) uint64 {
	return binary.BigEndian.Uint64(h[:8])
}
