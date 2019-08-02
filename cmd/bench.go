package cmd

import (
	"crypto/rand"
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
	hash.InitLXR()
	oprHash := make([]byte, 32)
	rand.Read(oprHash)

	// Instanciate SuperMiner
	miner := mining.NewSuperMiner(nbMiners)

	// Start miners
	log.Infof("Starting benchmarking for %s", duration)
	miner.Mine(oprHash, []byte{19, 89}, 3)
	ticker := time.NewTicker(duration)
	<-ticker.C
	miningSession := miner.Stop()

	// Print results
	log.Infof("Duration: %s", miningSession.Duration)
	log.Infof("Total ops: %d\n", miningSession.TotalOps)
	log.Infof("Hashrate: %d ops/s\n", uint64(float64(miningSession.TotalOps)/miningSession.Duration.Seconds()))
}
