package orax

import (
	"sync"
	"time"

	"github.com/pegnet/LXR256"
)

var LX lxr.LXRHash

type SuperMiner struct {
	running       bool
	miners        []*Miner
	wg            *sync.WaitGroup
	miningSession *MiningSession
}

type MiningSession struct {
	OprHash           []byte
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	TotalOps          uint64
	OrderedBestHashes []Hash
}

func NewSuperMiner(nbMiners int, maxBestHashes int) *SuperMiner {
	LX.Init(0x123412341234, 25, 256, 5)

	superMiner := new(SuperMiner)
	miners := make([]*Miner, nbMiners, nbMiners)
	superMiner.miners = miners

	for i := 0; i < nbMiners; i++ {
		miners[i] = NewMiner(i, maxBestHashes, &LX)
	}

	return superMiner
}

func (sm *SuperMiner) Mine(oprHash []byte) {
	if sm.running {
		panic("Tried to run an already running miner")
	}

	sm.running = true

	sm.miningSession = new(MiningSession)
	sm.miningSession.StartTime = time.Now()
	sm.miningSession.OprHash = oprHash

	wg := new(sync.WaitGroup)
	for i := 0; i < len(sm.miners); i++ {
		wg.Add(1)
		go sm.miners[i].mine(oprHash, wg)
	}
	sm.wg = wg
}

func (sm *SuperMiner) Stop() MiningSession {
	if !sm.running {
		panic("Tried to stop non-running miner")
	}

	for i := 0; i < len(sm.miners); i++ {
		sm.miners[i].stop <- 1
	}

	sm.wg.Wait()
	sm.running = false
	sm.miningSession.EndTime = time.Now()
	sm.miningSession.Duration = sm.miningSession.EndTime.Sub(sm.miningSession.StartTime)

	totalOps, orderedBestHashes := sm.computeMiningSessionResult()
	sm.miningSession.TotalOps = totalOps
	sm.miningSession.OrderedBestHashes = orderedBestHashes

	return *sm.miningSession
}

func (sm *SuperMiner) computeMiningSessionResult() (uint64, []Hash) {

	totalOps := uint64(0)
	var bestHashes []Hash
	for i := 0; i < len(sm.miners); i++ {
		totalOps += sm.miners[i].opsCounter
		bestHashes = append(bestHashes, sm.miners[i].bestHashes.diffOrderedHashes...)
	}
	sortHashesByDiff(bestHashes)

	return totalOps, bestHashes
}
