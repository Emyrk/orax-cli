package orax

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/rand"
	"time"

	flatbuffers "github.com/google/flatbuffers/go"
	"github.com/sirupsen/logrus"

	"gitlab.com/oraxpool/orax-cli/common"
	"gitlab.com/oraxpool/orax-message/msg"
	"gitlab.com/oraxpool/orax-message/msg/fbs"

	"gitlab.com/oraxpool/orax-cli/mining"
	"gitlab.com/oraxpool/orax-cli/ws"
)

var (
	log = common.GetLog()
)

type Client struct {
	wscli              *ws.Client
	miner              *mining.SuperMiner
	stopClaimingShares chan struct{}

	// Mining params
	CurrentTarget     uint64
	NoncePrefix       []byte
	InitialBatchDelay time.Duration
	BatchingDuration  time.Duration
}

type ClientConfig struct {
	NbMiners int
}

func (cli *Client) Start(config ClientConfig, stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})

	// Initialize super miner
	cli.miner = mining.NewSuperMiner(config.NbMiners)

	if common.GetIndicativeHashRate(config.NbMiners) == 0 {
		cli.benchHashRate()
	}

	// Initialize Websocket client
	cli.wscli = ws.NewWebSocketClient(config.NbMiners)

	go func() {
		defer close(done)
		cli.run(stop)
	}()

	return done
}

func (cli *Client) run(stop <-chan struct{}) {
	stopServer := make(chan struct{})
	doneServer := cli.wscli.Start(stopServer)

	for {
		select {
		case received, ok := <-cli.wscli.Receive:
			if !ok {
				return
			}
			cli.handleMessage(received)
		case coInfo, ok := <-cli.wscli.Connected:
			if !ok {
				return
			}

			// Populate connection information
			cli.NoncePrefix = coInfo.NoncePrefix
			cli.CurrentTarget = coInfo.Target
			cli.BatchingDuration = coInfo.BatchingDuration
			cli.InitialBatchDelay = coInfo.InitialBatchDelay
			log.WithField("params", coInfo).Info("Connected to Orax orchestrator")
		case _, ok := <-cli.wscli.Disconnected:
			if !ok {
				return
			}

			// If we lost the connection with the server
			// stop mining and claiming shares
			cli.stopClaimingShareBatches()
			if cli.miner.IsRunning() {
				cli.miner.Stop()
			}
		case <-stop:
			// Stop mining and send results
			cli.submitMiningResult(time.Duration(0))

			// Stop WS server
			close(stopServer)
			<-doneServer
			return
		}
	}
}

func (cli *Client) handleMessage(received []byte) {
	message, err := msg.UnmarshalMessage(received)
	if err != nil {
		log.WithError(err).Error("Failed to unmarshal message")
		return
	}

	switch v := message.(type) {
	case *fbs.StartMiningMessage:
		if cli.miner.IsRunning() {
			log.Warn("Stopping a stalled mining session")
			cli.miner.Stop()
		}
		cli.startClaimingShareBatches()
		cli.miner.Mine(v.OprHashBytes(), cli.NoncePrefix, cli.CurrentTarget)
	case *fbs.SubmissionWindowClosingMessage:
		cli.submitMiningResult(time.Duration(v.Deadline()) * time.Second)
	case *fbs.SetTargetMessage:
		cli.CurrentTarget = v.Target()
		log.Infof("New target set: %d", cli.CurrentTarget)
	default:
		log.Warnf("Unexpected message %T!\n", v)
	}
}

func (cli *Client) startClaimingShareBatches() {
	cli.stopClaimingShareBatches()

	cli.stopClaimingShares = make(chan struct{})
	go func() {
		timer := time.NewTimer(cli.InitialBatchDelay)
		select {
		case <-timer.C:
		case <-cli.stopClaimingShares:
			timer.Stop()
			return
		}

		cli.claimShareBatch()

		ticker := time.NewTicker(cli.BatchingDuration)
		for {
			select {
			case <-ticker.C:
				cli.claimShareBatch()
			case <-cli.stopClaimingShares:
				ticker.Stop()
				return
			}
		}
	}()
}

func (cli *Client) stopClaimingShareBatches() {
	if cli.stopClaimingShares != nil {
		close(cli.stopClaimingShares)
		cli.stopClaimingShares = nil
	}
}

func (cli *Client) claimShareBatch() {
	if cli.miner.IsRunning() {
		nonces := cli.miner.ReadNonceBuffer()
		if len(nonces) > 0 {
			data := msg.NewSubmitMessage(flatbuffers.NewBuilder(1024), nonces)
			select {
			case cli.wscli.Send <- data:
			default:
				log.Error("Skipping sending shares as Send channel is not available")
			}
		}
	}
}

func (cli *Client) submitMiningResult(windowDuration time.Duration) {
	cli.stopClaimingShareBatches()

	if cli.miner.IsRunning() {
		ms := cli.miner.Stop()

		// Flush residual nonces
		if len(ms.NonceBuffer) > 0 {
			// Randomly delay the reply within acceptable time window
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			jitter := time.Duration(int64(float64(windowDuration.Nanoseconds()) / 2 * r.Float64()))
			timer := time.NewTimer(jitter)
			<-timer.C

			data := msg.NewSubmitMessage(flatbuffers.NewBuilder(1024), ms.NonceBuffer)

			select {
			case cli.wscli.Send <- data:
			default:
				log.Error("Skipping sending mining result as Send channel is not available")
			}
		}

		logMiningSession(&ms)

		err := common.SaveIndicativeHashRate(cli.miner.SubMinerCount, ms.TotalOps, ms.Duration)
		if err != nil {
			log.WithError(err).Warn("Failed to save indicative hash rate")
		}
		log.Info("Waiting for next mining session...")
	}
}

func logMiningSession(ms *mining.MiningSession) {
	targetBuff := make([]byte, 8)
	binary.BigEndian.PutUint64(targetBuff, ms.Target)

	log.WithFields(logrus.Fields{
		"duration": ms.Duration,
		"shares":   ms.TotalShares,
		"target":   fmt.Sprintf("%x", targetBuff),
	}).Infof("End of mining session")
}

func (cli *Client) benchHashRate() {
	log.Info("Initial evaluation of miner hash rate")

	oprHash := make([]byte, 32)
	rand.Read(oprHash)

	cli.miner.Mine(oprHash, []byte{19, 89}, math.MaxUint64)
	timer := time.NewTimer(time.Minute)
	<-timer.C

	ms := cli.miner.Stop()

	err := common.SaveIndicativeHashRate(cli.miner.SubMinerCount, ms.TotalOps, ms.Duration)
	if err != nil {
		log.WithError(err).Fatal("Failed to save indicative hash rate")
	}
	log.Infof("Hash rate initially evaluated at %dh/s", common.GetIndicativeHashRate(cli.miner.SubMinerCount))
}
