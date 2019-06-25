package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	_log "gitlab.com/pbernier3/orax-cli/log"
	"gitlab.com/pbernier3/orax-cli/orax"

	"gitlab.com/pbernier3/orax-cli/mining"

	lxr "github.com/pegnet/LXR256"
)

var (
	LX lxr.LXRHash

	log = _log.New("main")
)

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	oraxCli := new(orax.Client)
	source := rand.NewSource(time.Now().UnixNano())
	rd := rand.New(source)
	id := strconv.Itoa(rd.Intn(100))
	oraxCli.Start("miner-" + id)

	sig := <-sigs
	log.Infof("[%s signal] Stopping Orax client... ", sig)
	oraxCli.Stop()

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
