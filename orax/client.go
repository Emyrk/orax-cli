package orax

import (
	"runtime"
	"sync"

	_log "gitlab.com/pbernier3/orax-cli/log"
	"gitlab.com/pbernier3/orax-cli/msg"

	"gitlab.com/pbernier3/orax-cli/mining"
	"gitlab.com/pbernier3/orax-cli/ws"
)

var (
	log = _log.New("orax")
)

type Client struct {
	wscli *ws.Client
	miner *mining.SuperMiner
}

func (cli *Client) Start(id string) {
	cli.wscli = new(ws.Client)
	go cli.wscli.Start(id)

	nbMiners := runtime.NumCPU()
	cli.miner = mining.NewSuperMiner(nbMiners)
	go cli.listenSignals()
}

func (cli *Client) Stop() {
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
	log.Info("Listening to signals")
	for {
		received, ok := <-cli.wscli.Received
		if !ok {
			return
		}

		message, err := msg.UnmarshalMessage(received)
		if err != nil {
			log.Warn("Failed to unmarshal message", err)
			continue
		}

		switch v := message.(type) {
		case *msg.MineSignalMessage:
			cli.miner.Mine(v.OprHash)
		case *msg.SubmitSignalMessage:
			if cli.miner.IsRunning() {
				ms := cli.miner.Stop()

				msm := msg.NewMinerSubmissionMessage()
				msm.OprHash = ms.OprHash
				msm.Nonce = ms.OrderedBestNonces[len(ms.OrderedBestNonces)-1].Nonce
				msm.Difficulty = ms.OrderedBestNonces[len(ms.OrderedBestNonces)-1].Difficulty
				msm.HashRate = uint64(float64(ms.TotalOps) / ms.Duration.Seconds())

				log.Infof("Submitting nonce [%x] for opr hash [%x] with difficult [%d].\nHashrate: [%d]",
					msm.Nonce, msm.OprHash, msm.Difficulty, msm.HashRate)

				cli.wscli.Send(msm.Marshal())
			}
		default:
			log.Warnf("Unexpected message %T!\n", v)
		}
	}
}
