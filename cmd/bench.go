package cmd

import (
	"crypto/rand"
	"fmt"
	"runtime"
	"time"

	"gitlab.com/pbernier3/orax-cli/hash"
	"gitlab.com/pbernier3/orax-cli/mining"

	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run a benchmark to evaluate the mining performance of the machine",
	Run: func(cmd *cobra.Command, args []string) {
		bench()
	},
}
var duration time.Duration

func init() {
	rootCmd.AddCommand(benchCmd)
	benchCmd.Flags().DurationVarP(&duration, "duration", "d", 1*time.Minute, "Duration of the benchmark")
	benchCmd.Flags().IntVarP(&nbMiners, "nbminer", "n", runtime.NumCPU(), "Number of concurrent miners. Default to number of logical CPUs. ")
}

func bench() {
	fmt.Printf("\nRunning benchmark for %s...\n\n", duration)

	hash.InitLXR()
	oprHash := make([]byte, 32)
	rand.Read(oprHash)

	// Instanciate SuperMiner
	miner := mining.NewSuperMiner(nbMiners)

	// Start miners
	miner.Mine(oprHash, []byte{19, 89}, 3)
	ticker := time.NewTicker(duration)
	<-ticker.C
	miningSession := miner.Stop()

	// Print results
	fmt.Printf("\n===================\n")
	fmt.Printf("Benchmarck results:\n")
	fmt.Printf("===================\n")
	fmt.Printf("%-15s %s\n", "Duration", miningSession.Duration)
	fmt.Printf("%-15s %d\n", "Total hashes", miningSession.TotalOps)
	fmt.Printf("%-15s %d hash/s\n", "Hash rate", uint64(float64(miningSession.TotalOps)/miningSession.Duration.Seconds()))
}
