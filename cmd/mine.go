package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
	"gitlab.com/oraxpool/orax-cli/hash"
	"gitlab.com/oraxpool/orax-cli/orax"
)

var nbMiners int

func init() {
	rootCmd.AddCommand(mineCmd)
	mineCmd.Flags().IntVarP(&nbMiners, "nbminer", "n", runtime.NumCPU(), "Number of concurrent miners. Default to number of logical CPUs.")
}

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "Start mining",
	Run: func(cmd *cobra.Command, args []string) {
		err := viper.ReadInConfig()

		if err != nil {
			color.Red("Failed to read config: %s", err)
		} else if viper.GetString("miner_id") == "" {
			fmt.Printf("\nTo start mining, first register your miner with the command `orax-cli register`\n\n")
		} else {
			os.Exit(mine())
		}
	},
}

func mine() int {
	hash.InitLXR()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	stopOraxCli := make(chan struct{})
	oraxCli := new(orax.Client)
	config := orax.ClientConfig{NbMiners: nbMiners}
	oraxCliDone := oraxCli.Start(config, stopOraxCli)

	if oraxCliDone == nil {
		return 1
	}

	defer func() {
		close(stopOraxCli) // Stop orax cli.
		fmt.Println("\n\nWaiting for Orax cli to stop...")
		<-oraxCliDone // Wait for orax cli to stop.
		color.Green("\nOrax cli stopped.\n\n")
	}()

	defer signal.Reset()
	// Wait for interrupt signal or unexpected termination of orax cli
	select {
	case <-sigs:
	case <-oraxCliDone: // Closed if Orax cli exits by itself (kicked by server).
	}

	return 0
}
