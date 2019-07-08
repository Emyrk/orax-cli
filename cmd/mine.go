package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"gitlab.com/pbernier3/orax-cli/common"

	"github.com/spf13/cobra"
	"gitlab.com/pbernier3/orax-cli/orax"
)

func init() {
	rootCmd.AddCommand(mineCmd)
}

var mineCmd = &cobra.Command{
	Use:   "mine",
	Short: "Start mining",
	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(_main())
	},
}

func _main() int {
	log := common.GetLog()

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
		log.WithField("signal", sig).Info("Interrupt signal received. Shutting down.")
		return 0
	case <-oraxCliDone: // Closed if Orax cli exits prematurely.
	}

	return 1
}
