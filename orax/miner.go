package orax

import (
	"crypto/rand"
	"encoding/binary"
	"sort"
	"sync"

	lxr "github.com/pegnet/LXR256"
)

type Miner struct {
	id            int
	hasher        *lxr.LXRHash
	maxBestHashes int
	stop          chan int
	opsCounter    uint64
	bestHashes    *BestHashes
}

func NewMiner(id int, maxBestHashes int, hasher *lxr.LXRHash) *Miner {
	miner := new(Miner)
	miner.id = id
	miner.hasher = hasher
	miner.maxBestHashes = maxBestHashes
	miner.stop = make(chan int, 1)
	miner.bestHashes = NewBestHashes(maxBestHashes)

	return miner
}

func (miner *Miner) Reset() {
	miner.opsCounter = 0
	miner.bestHashes = NewBestHashes(miner.maxBestHashes)
}

type BestHashes struct {
	maxLength         int
	diffOrderedHashes []Hash
}

type Hash struct {
	nonce      []byte
	difficulty uint64
}

func NewBestHashes(maxLength int) *BestHashes {
	bestHashes := new(BestHashes)
	bestHashes.maxLength = maxLength
	return bestHashes
}

func (bh *BestHashes) add(hash []byte, difficulty uint64) {
	if len(bh.diffOrderedHashes) == 0 || difficulty > bh.diffOrderedHashes[0].difficulty {
		bh.diffOrderedHashes = append(bh.diffOrderedHashes, Hash{hash, difficulty})
		sortHashesByDiff(bh.diffOrderedHashes)
		if len(bh.diffOrderedHashes) > bh.maxLength {
			bh.diffOrderedHashes = bh.diffOrderedHashes[1:]
		}
	}
}

func (miner *Miner) mine(oprHash []byte, wg *sync.WaitGroup) {
	nonce := make([]byte, 32)
	rand.Read(nonce)

mining:
	for i := 0; ; i++ {
		select {
		case <-miner.stop:
			break mining
		default:
		}

		miner.opsCounter++
		// TODO: write more clear?
		k := 0
		for j := i; j > 0; j = j >> 8 {
			nonce[k] = byte(j)
			k++
		}
		dataToHash := append(oprHash, nonce...)
		h := LX.Hash(dataToHash)
		diff := computeDifficulty(h)

		miner.bestHashes.add(nonce, diff)
	}
	wg.Done()
}

func computeDifficulty(h []byte) uint64 {
	return binary.BigEndian.Uint64(h[0:8])
}

func sortHashesByDiff(hashes []Hash) {
	sort.Slice(hashes, func(i, j int) bool {
		return hashes[i].difficulty < hashes[j].difficulty
	})
}
