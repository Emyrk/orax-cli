package orax

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"

	"gitlab.com/pbernier3/orax-cli/common"
	"gitlab.com/pbernier3/orax-cli/msg"

	"gitlab.com/pbernier3/orax-cli/mining"
	"gitlab.com/pbernier3/orax-cli/ws"
)

var (
	log = common.GetLog()
)

type Client struct {
	wscli *ws.Client
	miner *mining.SuperMiner
}

type ClientConfig struct {
	NbMiners int
}

func (cli *Client) Start(config ClientConfig, stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})

	// Initialize super miner
	cli.miner = mining.NewSuperMiner(config.NbMiners)

	// Initialize and start Websocket client
	cli.wscli = ws.NewWebSocketClient()

	go cli.wscli.Start()
	go cli.listenSignals()

	go func() {
		select {
		case <-stop:
			cli.stop()
			close(done)
		case <-done:
		}
	}()

	return done
}

func (cli *Client) stop() {
	// Try to submit immediately what the miner was working on
	// If we are actually connected to the orchestrator
	if cli.wscli.Connected() {
		cli.submitMiningResult(time.Duration(0))
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
			log.Error("Failed to unmarshal message: ", err)
			continue
		}

		switch v := message.(type) {
		case *msg.MineSignalMessage:
			if cli.miner.IsRunning() {
				log.Warn("Stopping a stalled mining session")
				cli.miner.Stop()
			}
			cli.miner.Mine(v.OprHash, cli.wscli.NoncePrefix, int(v.MaxNonces))
		case *msg.SubmitSignalMessage:
			cli.submitMiningResult(time.Duration(v.WindowDurationSec) * time.Second)
		default:
			log.Warnf("Unexpected message %T!\n", v)
		}
	}
}

func (cli *Client) submitMiningResult(windowDuration time.Duration) {
	if cli.miner.IsRunning() {
		ms := cli.miner.Stop()

		msm := new(msg.MinerSubmissionMessage)
		msm.OprHash = ms.OprHash

		msm.Nonces = make([]msg.Nonce, len(ms.OrderedBestNonces))
		for i, nonce := range ms.OrderedBestNonces {
			msm.Nonces[i] = msg.Nonce{Nonce: nonce.Nonce, Difficulty: nonce.Difficulty}
		}
		msm.OpCount = ms.TotalOps
		msm.Duration = ms.Duration.Nanoseconds()

		log.WithFields(logrus.Fields{
			"nonces":   msm.Nonces,
			"oprHash":  msm.OprHash,
			"opCount":  msm.OpCount,
			"hashRate": int64(float64(ms.TotalOps) / ms.Duration.Seconds()),
		}).Infof("Submitting mining result")

		data, err := msm.Marshal()
		if err != nil {
			log.Error("Failed to marshal MinerSubmissionMessage: ", err)
		} else {
			// Randomly delay the reply within acceptable time window
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			jitter := time.Duration(int64(float64(windowDuration.Nanoseconds()) / 2 * r.Float64()))
			timer := time.NewTimer(jitter)
			<-timer.C
			cli.wscli.Send(data)
		}
	}
}
