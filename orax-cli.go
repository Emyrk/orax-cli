package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_log "gitlab.com/pbernier3/orax-cli/log"
	"gitlab.com/pbernier3/orax-cli/orax"

	"gitlab.com/pbernier3/orax-cli/mining"

	lxr "github.com/pegnet/LXR256"
)

var (
	LX lxr.LXRHash

	log = _log.New("main")
)

func main() { os.Exit(_main()) }

func _main() int {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	stopOraxCli := make(chan struct{})
	oraxCli := new(orax.Client)
	oraxCliDone := oraxCli.Start(stopOraxCli)

	if oraxCliDone == nil {
		return 1
	}

	defer func() {
		close(stopOraxCli) // Stop orax cli.
		log.Info("Waiting for Orax cli to stop...")
		<-oraxCliDone // Wait for orax cli to stop.
		log.Info("Orax cli stopped.")
	}()

	defer signal.Reset()
	// Wait for interrupt signal or unexpected termination of orax cli
	select {
	case sig := <-sigs:
		log.Infof("%s signal received. Shutting down.", sig)
		return 0
	case <-oraxCliDone: // Closed if Orax cli exits prematurely.
	}

	return 1

	// oprHash := make([]byte, 32)
	// src := rand.NewSource(time.Now().UnixNano())
	// rd := rand.New(src)
	// rd.Read(oprHash)

	// // Instanciate SuperMiner
	// nbMiners := runtime.NumCPU()
	// miner := mining.NewSuperMiner(nbMiners)

	// fmt.Printf("Number of miners: %d\n", nbMiners)

	// // Start miners
	// miner.Mine(oprHash)
	// ticker := time.NewTicker(1 * time.Minute)
	// <-ticker.C
	// miningSession := miner.Stop()

	// printStats(miningSession)
}

func printStats(miningSession mining.MiningSession) {
	fmt.Println()
	fmt.Println("##############################")
	fmt.Println(miningSession.Duration)
	fmt.Printf("%d total ops\n", miningSession.TotalOps)
	fmt.Printf("%d ops/s\n", uint64(float64(miningSession.TotalOps)/miningSession.Duration.Seconds()))
}
