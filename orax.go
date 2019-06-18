package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	lxr "github.com/pegnet/LXR256"
	"gitlab.com/pbernier3/orax-cli/orax"
)

var LX lxr.LXRHash

func main() {
	LX.Init(0x123412341234, 25, 256, 5)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	oprHash := make([]byte, 32)
	rand.Read(oprHash)
	fmt.Println(oprHash)

	// Instanciate SuperMiner
	nbMiners := runtime.NumCPU()
	maxBestHashSize := 4
	miner := orax.NewSuperMiner(nbMiners, maxBestHashSize)

	fmt.Printf("Number of miners: %d\n", nbMiners)
	fmt.Printf("Keep max best hashes: %d\n", maxBestHashSize)

	// Start miners
	miner.Mine(oprHash)
	fmt.Println("Mining launched!")

	<-sigs
	fmt.Println()
	miningSession := miner.Stop()

	printStats(miningSession)
}

func printStats(miningSession orax.MiningSession) {
	fmt.Println("##############################")
	fmt.Println(miningSession.Duration)
	fmt.Printf("%d total ops\n", miningSession.TotalOps)
	fmt.Printf("%d ops/s\n", uint64(float64(miningSession.TotalOps)/miningSession.Duration.Seconds()))
	fmt.Println(len(miningSession.OrderedBestHashes))

}
