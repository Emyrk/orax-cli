package mining

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"gitlab.com/oraxpool/orax-cli/common"

	"github.com/sirupsen/logrus"
)

var log = common.GetLog()
var nonceBufferMux sync.Mutex

type SuperMiner struct {
	SubMinerCount int
	running       bool

	miners        []*Miner
	wg            *sync.WaitGroup
	miningSession *MiningSession
}

type MiningSession struct {
	NoncePrefix []byte
	OprHash     []byte
	StartTime   time.Time
	EndTime     time.Time
	Duration    time.Duration
	TotalOps    int64
	TotalShares int64
	NonceBuffer [][]byte
	sharesC     chan []byte
	Target      uint64
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

func (sm *SuperMiner) Mine(oprHash []byte, noncePrefix []byte, target uint64) {
	if sm.running {
		log.Fatal("Tried to run an already running miner")
	}

	sm.running = true

	sm.miningSession = new(MiningSession)

	sm.miningSession.Target = target
	sm.miningSession.NoncePrefix = noncePrefix
	sm.miningSession.StartTime = time.Now()
	sm.miningSession.OprHash = oprHash
	sm.miningSession.NonceBuffer = make([][]byte, 0, 300)

	sm.miningSession.sharesC = make(chan []byte, 64)
	go sm.collectShares(sm.miningSession.sharesC)

	wg := new(sync.WaitGroup)
	for i := 0; i < len(sm.miners); i++ {
		sm.miners[i].Reset()
		wg.Add(1)
		go sm.miners[i].mine(oprHash, noncePrefix, target, wg, sm.miningSession.sharesC)
	}
	sm.wg = wg

	targetBuff := make([]byte, 8)
	binary.BigEndian.PutUint64(targetBuff, target)

	log.WithFields(logrus.Fields{
		"nbSubMiners": len(sm.miners),
		"oprHash":     oprHash,
		"noncePrefix": noncePrefix,
		"target":      fmt.Sprintf("%x", targetBuff),
	}).Infof("Starting mining session")
}

func (sm *SuperMiner) collectShares(c <-chan []byte) {
	for nonce := range c {
		sm.miningSession.TotalShares++
		nonceBufferMux.Lock()
		sm.miningSession.NonceBuffer = append(sm.miningSession.NonceBuffer, nonce)
		nonceBufferMux.Unlock()
	}
}

func (sm *SuperMiner) ReadNonceBuffer() [][]byte {
	nonceBufferMux.Lock()
	buffer := sm.miningSession.NonceBuffer
	sm.miningSession.NonceBuffer = make([][]byte, 0, 300)
	nonceBufferMux.Unlock()
	return buffer
}

func (sm *SuperMiner) Stop() MiningSession {
	if !sm.running {
		log.Fatal("Tried to stop non-running miner")
	}

	for i := 0; i < len(sm.miners); i++ {
		sm.miners[i].stop <- 1
	}

	sm.wg.Wait()

	sm.running = false
	sm.miningSession.EndTime = time.Now()
	sm.miningSession.Duration = sm.miningSession.EndTime.Sub(sm.miningSession.StartTime)
	close(sm.miningSession.sharesC)

	for i := 0; i < len(sm.miners); i++ {
		sm.miningSession.TotalOps += sm.miners[i].opsCounter
	}

	return *sm.miningSession
}

func (sm *SuperMiner) IsRunning() bool {
	return sm.running
}
