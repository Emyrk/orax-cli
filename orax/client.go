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
	"gitlab.com/oraxpool/orax-cli/miningv0"
	"gitlab.com/oraxpool/orax-cli/ws"
)

var (
	log = common.GetLog()
)

type Client struct {
	wscli              *ws.Client
	miner              *mining.SuperMiner
	stopClaimingShares chan struct{}
	minerV0            *miningv0.SuperMiner
	// TODO: delete once swithced to v1
	miningVersion byte

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
	cli.minerV0 = miningv0.NewSuperMiner(config.NbMiners)

	if common.GetIndicativeHashRate(config.NbMiners) == 0 {
		cli.benchHashRate()
	}

	// Initialize and start Websocket client
	cli.wscli = ws.NewWebSocketClient(config.NbMiners)

	go cli.onConnected()
	go cli.wscli.Start()
	go cli.listenSignals()

	go func() {
		select {
		case <-stop:
			cli.stop()
			close(done)
		case <-cli.wscli.DoneReading:
			close(done)
		}
	}()

	return done
}

func (cli *Client) onConnected() {
	for coInfo := range cli.wscli.Connected {
		cli.NoncePrefix = coInfo.NoncePrefix
		cli.CurrentTarget = coInfo.Target
		cli.BatchingDuration = coInfo.BatchingDuration
		cli.InitialBatchDelay = coInfo.InitialBatchDelay
		if coInfo.Target > 0 {
			cli.miningVersion = 1
		} else {
			cli.miningVersion = 0
		}

		log.WithField("params", coInfo).Info("Connected to Orax orchestrator")
	}
}

func (cli *Client) stop() {
	// Try to submit immediately what the miner was working on
	// If we are actually connected to the orchestrator
	if cli.wscli.IsConnected() {
		if cli.miningVersion == 0 {
			cli.submitMiningResultV0(time.Duration(0))
		} else {
			cli.submitMiningResult(time.Duration(0))
		}
	}
	// Stop the webserver
	cli.wscli.Stop()
}

func (cli *Client) listenSignals() {
	for {
		received, ok := <-cli.wscli.Received
		if !ok {
			return
		}

		message, err := msg.UnmarshalMessage(received)
		if err != nil {
			log.WithError(err).Error("Failed to unmarshal message")
			continue
		}

		switch v := message.(type) {

		// V1 messages
		case *fbs.StartMiningMessage:
			cli.miningVersion = 1
			if cli.miner.IsRunning() {
				log.Warn("Stopping a stalled mining session")
				cli.miner.Stop()
			}
			cli.startClaimingShareBatches()
			cli.miner.Mine(v.OprHashBytes(), cli.NoncePrefix, cli.CurrentTarget)
		case *fbs.SubmissionWindowClosingMessage:
			cli.miningVersion = 1
			cli.submitMiningResult(time.Duration(v.Deadline()) * time.Second)
		case *fbs.SetTargetMessage:
			cli.miningVersion = 1
			cli.CurrentTarget = v.Target()
			log.Infof("New target set: %d", cli.CurrentTarget)

		// V0 messages
		case *msg.MineSignalMessage:
			cli.miningVersion = 0
			if cli.minerV0.IsRunning() {
				log.Warn("Stopping a stalled mining session")
				cli.minerV0.Stop()
			}
			cli.minerV0.Mine(v.OprHash, cli.NoncePrefix, int(v.MaxNonces))
		case *msg.SubmitSignalMessage:
			cli.miningVersion = 0
			cli.submitMiningResultV0(time.Duration(v.WindowDurationSec) * time.Second)
		default:
			log.Warnf("Unexpected message %T!\n", v)
		}
	}
}

///////////////
// V1
///////////////

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
	if cli.miner.IsRunning() && cli.wscli.IsConnected() {
		nonces := cli.miner.ReadNonceBuffer()
		if len(nonces) > 0 {
			data := msg.NewSubmitMessage(flatbuffers.NewBuilder(1024), nonces)
			cli.wscli.Send(data)
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

			cli.wscli.Send(data)
		}

		logMiningSession(&ms)

		err := common.SaveIndicativeHashRate(cli.miner.SubMinerCount, ms.TotalOps, ms.Duration)
		if err != nil {
			log.WithError(err).Warn("Failed to save indicative hash rate")
		}
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

///////////////
// V0
///////////////

func (cli *Client) submitMiningResultV0(windowDuration time.Duration) {
	if cli.minerV0.IsRunning() {
		ms := cli.minerV0.Stop()

		msm := new(msg.MinerSubmissionMessage)
		msm.OprHash = ms.OprHash

		msm.Nonces = make([]msg.Nonce, len(ms.OrderedBestNonces))
		for i, nonce := range ms.OrderedBestNonces {
			msm.Nonces[i] = msg.Nonce{Nonce: nonce.Nonce, Difficulty: nonce.Difficulty}
		}
		msm.OpCount = ms.TotalOps
		msm.Duration = ms.Duration.Nanoseconds()

		data, err := msm.Marshal()
		if err != nil {
			log.WithError(err).Error("Failed to marshal MinerSubmissionMessage")
		} else {
			// Randomly delay the reply within acceptable time window
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			jitter := time.Duration(int64(float64(windowDuration.Nanoseconds()) / 2 * r.Float64()))
			timer := time.NewTimer(jitter)
			<-timer.C
			cli.wscli.Send(data)

			logMiningResult(&ms)
		}

		err = common.SaveIndicativeHashRate(cli.minerV0.SubMinerCount, ms.TotalOps, ms.Duration)
		if err != nil {
			log.WithError(err).Warn("Failed to save indicative hash rate")
		}
	}
}

func logMiningResult(ms *miningv0.MiningSession) {
	nonces := make([]struct {
		Nonce      []byte
		Difficulty string
	}, len(ms.OrderedBestNonces))

	for i, nonce := range ms.OrderedBestNonces {
		diffBuff := make([]byte, 8)
		binary.BigEndian.PutUint64(diffBuff, nonce.Difficulty)
		nonces[i].Nonce = nonce.Nonce
		nonces[i].Difficulty = fmt.Sprintf("%x", diffBuff)
	}

	log.WithFields(logrus.Fields{
		"nonces":   nonces,
		"oprHash":  ms.OprHash,
		"hashRate": int64(float64(ms.TotalOps) / ms.Duration.Seconds()),
	}).Infof("Submitting mining result")
}
