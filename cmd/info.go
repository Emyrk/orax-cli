package cmd

import (
	"fmt"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/fatih/color"
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
			fmt.Printf("\n")
			color.Red(err.Error())
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
		fmt.Printf("\nLog in:\n\n")
		userID, jwt, err = existingOraxUser()
		if err != nil {
			return err
		}
		viper.Set("user_id", userID)
		viper.Set("jwt", jwt)

		fmt.Printf("\n")
	}

	userInfo, err := api.GetUserInfo(userID)
	if err != nil {
		if err == api.ErrAuth {
			// The JWT may be expired, prompt for credentials
			fmt.Printf("\nLog in:\n\n")
			userID, jwt, err = existingOraxUser()
			if err != nil {
				return err
			}

			viper.Set("user_id", userID)
			viper.Set("jwt", jwt)

			fmt.Printf("\n")

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

	fmt.Printf("==============================================================================\n")
	fmt.Printf("%-22s %s\n", "UserID", userID)
	fmt.Printf("%-22s %s\n", "Email", userInfo.User.Email)
	fmt.Printf("%-22s %s\n", "Registration Date", userInfo.User.RegistrationDate.Format(time.RFC3339))
	fmt.Printf("%-22s %s\n", "Payout Address", userInfo.User.PayoutAddress)
	fmt.Printf("%-22s %s PEG\n", "Balance", humanize.CommafWithDigits(userInfo.User.Balance/1e8, 8))
	fmt.Printf("==============================================================================\n")

	// Miners
	fmt.Printf("\nMiners:\n\n")

	minersTable := tablewriter.NewWriter(os.Stdout)
	minersTable.SetHeader([]string{"", "Alias", "Registration date", "Latest instant hashrate", "Latest submission"})

	minersTableData := make([][]string, len(userInfo.Miners))
	for i, miner := range userInfo.Miners {
		minersTableData[i] = []string{
			fmt.Sprintf("%d", i+1),
			miner.Alias,
			miner.RegistrationDate.Format(time.RFC3339),
			getHashRate(miner.LatestOpCount, miner.LatestDuration),
			humanize.Time(miner.LatestSubmissionDate),
		}
	}

	minersTable.AppendBulk(minersTableData)
	minersTable.Render()

	// Blocks stats
	fmt.Printf("\nLatest stats:\n\n")

	statsTable := tablewriter.NewWriter(os.Stdout)
	statsTable.SetAlignment(tablewriter.ALIGN_RIGHT)
	statsTable.SetHeader([]string{"Block", "Miners", "Pool hashrate", "Pool Reward", "User hashrate", "User share", "User reward"})

	statsTableData := make([][]string, len(userInfo.Stats))
	for i, stat := range userInfo.Stats {
		statsTableData[i] = []string{
			fmt.Sprintf("%s", humanize.Comma(int64(stat.Height))),
			fmt.Sprintf("%s", humanize.Comma(int64(stat.MinerCount))),
			fmt.Sprintf("%s", getHashRate(stat.TotalOpCount, stat.Duration)),
			fmt.Sprintf("%s", humanize.Commaf(float64(stat.UsersReward)/1e8)),
		}

		if stat.UserDetail != nil {
			statsTableData[i] = append(statsTableData[i],
				fmt.Sprintf("%s", getHashRate(stat.UserDetail.OpCount, stat.Duration)),
				fmt.Sprintf("%s%%", humanize.FtoaWithDigits(stat.UserDetail.Share*100, 2)),
				fmt.Sprintf("%s", humanize.CommafWithDigits(stat.UserDetail.Reward/1e8, 8)))
		} else {
			statsTableData[i] = append(statsTableData[i], "0", "0%", "0")
		}
	}
	statsTable.AppendBulk(statsTableData)
	statsTable.Render()

	return nil
}

func getHashRate(opCount int64, duration int64) string {
	if duration == 0 {
		return "0"
	}

	hashrate := int64(float64(opCount) / (float64(duration) / 1e9))
	return humanize.Comma(hashrate)
}
