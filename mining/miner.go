package mining

import (
	"encoding/binary"
	"sync"

	"gitlab.com/oraxpool/orax-cli/hash"
)

type Miner struct {
	id         int
	stop       chan int
	opsCounter int64
}

func NewMiner(id int) *Miner {
	miner := new(Miner)
	miner.id = id
	miner.stop = make(chan int)

	return miner
}

func (miner *Miner) Reset() {
	miner.opsCounter = 0
}

func (miner *Miner) mine(oprHash []byte, noncePrefix []byte, target uint64, wg *sync.WaitGroup, c chan<- []byte) {
	// Create a slice of sufficient capacity to avoid a new underlying array to be allocated
	// when appending nonce after the OPR
	dataToMine := make([]byte, 32, 64)
	copy(dataToMine, oprHash)

	prefixLength := len(noncePrefix) + 1
	// Pre allocate a large enough slice of memory
	nonce := make([]byte, 0, 64)
	// Append the noncePrefix of the super miner, the local prefix (miner id) and the first 0
	nonce = append(nonce, noncePrefix...)
	nonce = append(nonce, byte(miner.id), 0)

mining:
	for {
		// Listen for end of mining signal
		select {
		case <-miner.stop:
			break mining
		default:
		}

		// Increment nonce
		i := prefixLength
		for {
			nonce[i]++
			// if it overflows
			if nonce[i] == 0 {
				i++
				// If we reached the end of the slice, expand it
				if i == len(nonce) {
					nonce = append(nonce, 0)
					break
				}
			} else {
				break
			}
		}

		// Compute hash and difficulty
		dataToHash := append(dataToMine, nonce...)
		h := hash.Hash(dataToHash)
		diff := computeDifficulty(h)
		miner.opsCounter++

		if diff >= target {
			c <- copyNonce(nonce)
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
