package orax

import (
	"runtime"
	"sync"

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

func (cli *Client) Start(stop <-chan struct{}) <-chan struct{} {
	done := make(chan struct{})

	// Initialize super miner
	nbMiners := runtime.NumCPU()
	cli.miner = mining.NewSuperMiner(nbMiners)

	// Initialize and start Websocket client
	cli.wscli = new(ws.Client)
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
	wg := new(sync.WaitGroup)

	wg.Add(1)
	go func() {
		cli.wscli.Stop()
		wg.Done()
	}()

	if cli.miner.IsRunning() {
		wg.Add(1)
		go func() {
			cli.miner.Stop()
			wg.Done()
		}()
	}

	wg.Wait()
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
			cli.miner.Mine(v.OprHash)
		case *msg.SubmitSignalMessage:
			if cli.miner.IsRunning() {
				ms := cli.miner.Stop()

				msm := msg.NewMinerSubmissionMessage()
				msm.OprHash = ms.OprHash
				msm.Nonce = ms.OrderedBestNonces[len(ms.OrderedBestNonces)-1].Nonce
				msm.Difficulty = ms.OrderedBestNonces[len(ms.OrderedBestNonces)-1].Difficulty
				msm.HashRate = uint64(float64(ms.TotalOps) / ms.Duration.Seconds())

				log.WithFields(logrus.Fields{
					"nonce":      msm.Nonce,
					"oprHash":    msm.OprHash,
					"difficulty": msm.Difficulty,
					"hashRate":   msm.HashRate,
				}).Infof("Submitting mining result")

				cli.wscli.Send(msm.Marshal())
			} else {
				log.Info("Skipping submission as no mining session was running.")
			}
		default:
			log.Warnf("Unexpected message %T!\n", v)
		}
	}
}
