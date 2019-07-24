package mining

import (
	"sort"
	"sync"
	"time"

	"gitlab.com/pbernier3/orax-cli/common"

	"github.com/sirupsen/logrus"
)

var log = common.GetLog()

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
	OrderedBestNonces []Nonce
}

func NewSuperMiner(nbMiners int) *SuperMiner {
	superMiner := new(SuperMiner)
	miners := make([]*Miner, nbMiners, nbMiners)
	superMiner.miners = miners

	for i := 0; i < nbMiners; i++ {
		miners[i] = NewMiner(i)
	}

	return superMiner
}

func (sm *SuperMiner) Mine(oprHash []byte) {
	if sm.running {
		log.Fatal("Tried to run an already running miner")
	}

	log.WithFields(logrus.Fields{
		"nbSubMiners": len(sm.miners),
		"oprHash":     oprHash,
	}).Infof("Starting mining")

	sm.running = true

	sm.miningSession = new(MiningSession)
	sm.miningSession.StartTime = time.Now()
	sm.miningSession.OprHash = oprHash

	wg := new(sync.WaitGroup)
	for i := 0; i < len(sm.miners); i++ {
		sm.miners[i].Reset()
		wg.Add(1)
		go sm.miners[i].mine(oprHash, wg)
	}
	sm.wg = wg
}

func (sm *SuperMiner) Stop() MiningSession {
	if !sm.running {
		log.Fatal("Tried to stop non-running miner")
	}
	log.Info("Stopping mining")

	for i := 0; i < len(sm.miners); i++ {
		sm.miners[i].stop <- 1
	}

	sm.wg.Wait()
	sm.running = false
	sm.miningSession.EndTime = time.Now()
	sm.miningSession.Duration = sm.miningSession.EndTime.Sub(sm.miningSession.StartTime)

	totalOps, orderedBestNonces := sm.computeMiningSessionResult()
	sm.miningSession.TotalOps = totalOps
	sm.miningSession.OrderedBestNonces = orderedBestNonces

	return *sm.miningSession
}

func (sm *SuperMiner) computeMiningSessionResult() (uint64, []Nonce) {

	totalOps := uint64(0)
	var bestNonces []Nonce
	for i := 0; i < len(sm.miners); i++ {
		totalOps += sm.miners[i].opsCounter

		// bestNonce can be nil if the miner didn't have the time to compute a single hash
		if sm.miners[i].bestNonce != nil {
			bestNonces = append(bestNonces, *sm.miners[i].bestNonce)
		}
	}
	sortNoncesByDiff(bestNonces)

	return totalOps, bestNonces
}

func sortNoncesByDiff(nonces []Nonce) {
	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i].Difficulty < nonces[j].Difficulty
	})
}

func (sm *SuperMiner) IsRunning() bool {
	return sm.running
}
