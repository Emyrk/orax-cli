package cmd

import (
	"fmt"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/pbernier3/orax-cli/api"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get miner info",
	Run: func(cmd *cobra.Command, args []string) {
		viper.ReadInConfig()
		err := info()
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}

func info() (err error) {
	userID := viper.GetString("user_id")
	jwt := viper.GetString("jwt")

	if jwt == "" || userID == "" {
		fmt.Println("Log in:")
		userID, jwt, err = existingOraxUser()
		if err != nil {
			return err
		}
		viper.Set("user_id", userID)
		viper.Set("jwt", jwt)
	}

	userInfo, err := api.GetUserInfo(userID)
	if err != nil {
		if err == api.ErrAuth {
			// The JWT may be expired, prompt for credentials
			fmt.Println("Log in:")
			userID, jwt, err = existingOraxUser()
			if err != nil {
				return err
			}

			viper.Set("user_id", userID)
			viper.Set("jwt", jwt)

			userInfo, err = api.GetUserInfo(userID)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	// Save the fresh JWT token for next time
	err = viper.WriteConfig()
	if err != nil {
		return err
	}

	fmt.Printf("\n%-20s %s\n", "UserID", userID)
	fmt.Printf("%-20s %s\n", "Email", userInfo.User.Email)
	fmt.Printf("%-20s %s\n", "Registration Date", userInfo.User.RegistrationDate.Format(time.RFC3339))
	fmt.Printf("%-20s %s\n", "Payout Address", userInfo.User.PayoutAddress)
	fmt.Printf("%-20s %s PNT\n", "Balance", humanize.CommafWithDigits(userInfo.User.Balance/1e8, 8))
	fmt.Printf("\nMiners:\n\n")
	// Miners
	minersTableData := make([][]string, len(userInfo.Miners))
	for i, miner := range userInfo.Miners {
		minersTableData[i] = []string{
			fmt.Sprintf("%d", i+1),
			miner.Alias,
			miner.RegistrationDate.Format(time.RFC3339),
			humanize.Comma(int64(miner.LatestHashRate)),
			miner.LatestSubmissionDate.Format(time.RFC3339),
		}
	}
	minersTable := tablewriter.NewWriter(os.Stdout)
	minersTable.SetHeader([]string{"", "Alias", "Registration date", "Latest hashrate", "Latest submission date"})
	minersTable.AppendBulk(minersTableData)
	minersTable.Render()

	// Blocks
	fmt.Printf("\nLatest stats:\n\n")
	statsTableData := make([][]string, len(userInfo.Stats))
	for i, stat := range userInfo.Stats {
		statsTableData[i] = []string{
			fmt.Sprintf("%s", humanize.Comma(int64(stat.Height))),
			fmt.Sprintf("%s", humanize.Comma(int64(stat.NbMiners))),
			fmt.Sprintf("%s", humanize.Comma(int64(stat.TotalHashRate))),
			fmt.Sprintf("%s", humanize.Commaf(float64(stat.UsersReward)/1e8)),
		}

		if stat.UserDetail != nil {
			statsTableData[i] = append(statsTableData[i],
				fmt.Sprintf("%s", humanize.Comma(int64(stat.UserDetail.HashRate))),
				fmt.Sprintf("%s%%", humanize.FtoaWithDigits(stat.UserDetail.Share*100, 2)),
				fmt.Sprintf("%s", humanize.CommafWithDigits(stat.UserDetail.Reward/1e8, 8)))
		} else {
			statsTableData[i] = append(statsTableData[i], "0", "0%", "0")
		}
	}
	statsTable := tablewriter.NewWriter(os.Stdout)
	statsTable.SetAlignment(tablewriter.ALIGN_RIGHT)
	statsTable.SetHeader([]string{"Height", "Miners", "Pool hashrate", "Pool Reward", "User hashrate", "User share", "User reward"})
	statsTable.AppendBulk(statsTableData)
	statsTable.Render()

	return nil
}
