package mining

import (
	"encoding/binary"
	"fmt"
	"sync"

	lxr "github.com/pegnet/LXRHash"
	"gitlab.com/oraxpool/orax-cli/hash"
)

var _ = fmt.Printf
var _ = lxr.Init

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

	//prefixLength := len(noncePrefix) + 1
	// Pre allocate a large enough slice of memory
	nonce := make([]byte, 0, 64)
	// Append the noncePrefix of the super miner, the local prefix (miner id) and the first 0
	nonce = append(nonce, noncePrefix...)
	nonce = append(nonce, byte(miner.id))
	ni := NewNonceIncrementer(nonce)

	static := append(oprHash, append(noncePrefix, byte(miner.id))...)
	//cache := hash.LX.Cache(static)
	//var _ = cache

	//func miningRoutine(lx *lxr.LXRHash, base []byte, id byte, batchsize int) {
	//	var start uint32
	//	batch := make([][]byte, batchsize)
	//	buf := make([]byte, 4)
	//	for {
	//		for i := range batch {
	//			batch[i] = []byte{id}
	//			binary.BigEndian.PutUint32(buf, start+uint32(i))
	//			batch[i] = append(batch[i], buf...)
	//		}
	//		start += uint32(batchsize)
	//
	//		results := lx.HashWork(base, batch)
	//		for i := range results {
	//			// do something with the result here
	//			// nonce = batch[i]
	//			// input = append(base, batch[i]...)
	//			// hash = results[i]
	//		}
	//	}
	//}

	if false {
	mining:
		for {
			// Listen for end of mining signal
			select {
			case <-miner.stop:
				break mining
			default:
			}

			// Increment nonce
			//i := prefixLength
			//for {
			//	nonce[i]++
			//	// if it overflows
			//	if nonce[i] == 0 {
			//		i++
			//		// If we reached the end of the slice, expand it
			//		if i == len(nonce) {
			//			nonce = append(nonce, 0)
			//			break
			//		}
			//	} else {
			//		break
			//	}
			//}

			changed := ni.NextNonce()

			// Compute hash and difficulty
			dataToHash := append(dataToMine, ni.Nonce...)
			if changed == -1 {
				//fmt.Println("cached")
				//cache = hash.LX.Cache(dataToHash[:len(dataToHash)-1])
			}
			//h := cache.Hash(dataToHash)
			h := hash.LX.Hash(dataToHash)
			diff := computeDifficulty(h)
			miner.opsCounter++

			if diff >= target {
				c <- copyNonce(ni.Nonce)
			}
		}

		wg.Done()
	} else {
		var start uint32
	miningBatch:
		for {
			// Listen for end of mining signal
			select {
			case <-miner.stop:
				break miningBatch
			default:
			}

			batchsize := 256
			batch := make([][]byte, batchsize)

			for i := range batch {
				batch[i] = make([]byte, 4)
				binary.BigEndian.PutUint32(batch[i], start+uint32(i))
			}
			start += uint32(batchsize)

			results := hash.LX.HashWork(static, batch)
			for i := range results {
				// do something with the result here
				// nonce = batch[i]
				// input = append(base, batch[i]...)
				// hash = results[i]
				h := results[i]
				diff := computeDifficulty(h)
				miner.opsCounter++

				if diff >= target {
					c <- copyNonce(append(static[32:], batch[i]...))
				}
			}

			//changed := ni.NextNonce()
			//
			//// Compute hash and difficulty
			//dataToHash := append(dataToMine, ni.Nonce...)
			//if changed == -1 {
			//	//fmt.Println("cached")
			//	//cache = hash.LX.Cache(dataToHash[:len(dataToHash)-1])
			//}
			////h := cache.Hash(dataToHash)
			//h := hash.LX.Hash(dataToHash)
			//diff := computeDifficulty(h)
			//miner.opsCounter++
			//
			//if diff >= target {
			//	c <- copyNonce(ni.Nonce)
			//}
		}

		wg.Done()
	}
}

// NonceIncrementer is just simple to increment nonces
type NonceIncrementer struct {
	Nonce          []byte
	lastNonceByte  int
	lastPrefixByte int
}

func NewNonceIncrementer(prefix []byte) *NonceIncrementer {
	n := new(NonceIncrementer)

	n.lastPrefixByte = len(prefix) - 1
	n.Nonce = append(prefix, 0)
	n.lastNonceByte = 1
	return n
}

// NextNonce is just counting to get the next nonce. We preserve
// the first byte, as that is our ID and give us our nonce space
//	So []byte(ID, 255) -> []byte(ID, 1, 0) -> []byte(ID, 1, 1)
func (i *NonceIncrementer) NextNonce() int {
	idx := len(i.Nonce) - i.lastNonceByte
	for {
		i.Nonce[idx]++
		if i.Nonce[idx] == 0 {
			idx--
			if idx == i.lastPrefixByte { // This is my prefix, don't touch it!
				rest := append([]byte{1}, i.Nonce[i.lastPrefixByte+1:]...)
				i.Nonce = append(i.Nonce[:i.lastPrefixByte+1], rest...)
				return -1
			}
		} else {
			break
		}
	}
	return idx
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
