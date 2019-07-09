package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/pbernier3/orax-cli/common"
)

var rootCmd = &cobra.Command{
	Use:   "orax",
	Short: "Mining client for the Orax mining pool",
}

var log = common.GetLog()

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
