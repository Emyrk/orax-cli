package miningv0

import (
	"sort"
	"sync"
	"time"

	"gitlab.com/oraxpool/orax-cli/common"

	"github.com/sirupsen/logrus"
)

var log = common.GetLog()

type SuperMiner struct {
	SubMinerCount int

	running   bool
	maxNonces int

	miners        []*Miner
	wg            *sync.WaitGroup
	miningSession *MiningSession
}

type MiningSession struct {
	NoncePrefix       []byte
	OprHash           []byte
	StartTime         time.Time
	EndTime           time.Time
	Duration          time.Duration
	TotalOps          int64
	OrderedBestNonces []*Nonce
}

func NewSuperMiner(nbMiners int) *SuperMiner {
	superMiner := new(SuperMiner)
	superMiner.SubMinerCount = nbMiners
	miners := make([]*Miner, nbMiners, nbMiners)
	superMiner.miners = miners

	for i := 0; i < nbMiners; i++ {
		miners[i] = NewMiner(i)
	}

	return superMiner
}

func (sm *SuperMiner) Mine(oprHash []byte, noncePrefix []byte, maxNonces int) {
	if sm.running {
		log.Fatal("Tried to run an already running miner")
	}
	if maxNonces <= 0 {
		log.WithField("maxNonces", maxNonces).Error("Invalid maxNonces")
		return
	}

	log.WithFields(logrus.Fields{
		"nbSubMiners": len(sm.miners),
		"oprHash":     oprHash,
		"noncePrefix": noncePrefix,
		"maxNonces":   maxNonces,
	}).Infof("Starting mining")

	sm.running = true
	sm.maxNonces = maxNonces

	sm.miningSession = new(MiningSession)
	sm.miningSession.NoncePrefix = noncePrefix
	sm.miningSession.StartTime = time.Now()
	sm.miningSession.OprHash = oprHash

	wg := new(sync.WaitGroup)
	for i := 0; i < len(sm.miners); i++ {
		sm.miners[i].Reset()
		wg.Add(1)
		go sm.miners[i].mine(oprHash, noncePrefix, maxNonces, wg)
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

	if sm.maxNonces < len(orderedBestNonces) {
		sm.miningSession.OrderedBestNonces = orderedBestNonces[:sm.maxNonces]
	} else {
		sm.miningSession.OrderedBestNonces = orderedBestNonces
	}

	return *sm.miningSession
}

func (sm *SuperMiner) computeMiningSessionResult() (int64, []*Nonce) {

	totalOps := int64(0)
	var bestNonces []*Nonce
	for i := 0; i < len(sm.miners); i++ {
		totalOps += sm.miners[i].opsCounter
		bestNonces = append(bestNonces, sm.miners[i].bestNonces...)
	}
	SortNoncesByDiff(bestNonces)

	return totalOps, bestNonces
}

func SortNoncesByDiff(nonces []*Nonce) {
	sort.Slice(nonces, func(i, j int) bool {
		return nonces[i].Difficulty > nonces[j].Difficulty
	})
}

func (sm *SuperMiner) IsRunning() bool {
	return sm.running
}
