package cmd

import (
	"fmt"
	"os"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/oraxpool/orax-cli/api"
	"gitlab.com/oraxpool/orax-cli/common"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Get miner info",
	Run: func(cmd *cobra.Command, args []string) {
		viper.ReadInConfig()
		err := info()
		if err != nil {
			common.PrintError("%s\n", err.Error())
			os.Exit(1)
		}
	},
}
var (
	startHeight int
	limit       int
)

func init() {
	rootCmd.AddCommand(infoCmd)
	infoCmd.Flags().IntVarP(&startHeight, "start-height", "s", 0, "Height to start to retrieve block stats at.")
	infoCmd.Flags().IntVarP(&limit, "limit", "l", 18, "Number of blocks to retrieve statistics about.")
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

	userInfo, err := api.GetUserInfo(userID, startHeight, limit)
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

			userInfo, err = api.GetUserInfo(userID, startHeight, limit)
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
	fmt.Printf("%-22s %s PEG\n", "Total Reward", humanize.CommafWithDigits(userInfo.User.TotalReward/1e8, 8))
	fmt.Printf("%-22s %s PEG\n", "Pending Payment", humanize.CommafWithDigits(userInfo.User.Balance/1e8, 8))
	fmt.Printf("==============================================================================\n")

	// Miners
	fmt.Printf("\nMiners:\n\n")

	minersTable := tablewriter.NewWriter(os.Stdout)
	minersTable.SetHeader([]string{"", "Alias", "Registration date", "Latest effective hash rate", "Latest block participation"})

	minersTableData := make([][]string, len(userInfo.Miners))
	for i, miner := range userInfo.Miners {
		minersTableData[i] = []string{
			fmt.Sprintf("%d", i+1),
			miner.Alias,
			miner.RegistrationDate.Format(time.RFC3339),
		}

		// New effective hash rate system
		if miner.LatestEffectiveOpCount > 0 {
			minersTableData[i] = append(minersTableData[i], getHashRate(miner.LatestEffectiveOpCount, miner.LatestDuration))
		} else {
			// Old reported hash rate
			minersTableData[i] = append(minersTableData[i], getHashRate(miner.LatestOpCount, miner.LatestDuration))
		}

		minersTableData[i] = append(minersTableData[i], humanize.Comma(miner.LatestSubmissionHeight))
	}

	minersTable.AppendBulk(minersTableData)
	minersTable.Render()

	// Blocks stats
	fmt.Printf("\nLatest stats:\n\n")

	statsTable := tablewriter.NewWriter(os.Stdout)
	statsTable.SetAlignment(tablewriter.ALIGN_RIGHT)
	statsTable.SetHeader([]string{"Block", "Miners", "Pool scoring hash rate", "Pool Users Reward", "User scoring hash rate", "User share", "User reward"})

	statsTableData := make([][]string, len(userInfo.Stats))
	for i, stat := range userInfo.Stats {
		statsTableData[i] = []string{
			fmt.Sprintf("%s", humanize.Comma(int64(stat.Height))),
			fmt.Sprintf("%s", humanize.Comma(int64(stat.MinerCount))),
		}

		// New scoring system
		if stat.TotalScore > 0 {
			statsTableData[i] = append(statsTableData[i], humanize.CommafWithDigits(stat.TotalScore, 0))
		} else {
			// DEPRECATED: Old op count system
			statsTableData[i] = append(statsTableData[i], getHashRate(stat.TotalOpCount, stat.MiningDuration))
		}

		statsTableData[i] = append(statsTableData[i], fmt.Sprintf("%s", humanize.Commaf(float64(stat.UsersReward)/1e8)))

		if stat.UserDetail != nil {

			// New scoring system
			if stat.UserDetail.Score > 0 {
				statsTableData[i] = append(statsTableData[i],
					fmt.Sprintf("%s", humanize.CommafWithDigits(stat.UserDetail.Score, 0)))
			} else {
				// DEPRECATED: Old op count system
				statsTableData[i] = append(statsTableData[i],
					fmt.Sprintf("%s", getHashRate(stat.UserDetail.OpCount, stat.MiningDuration)))
			}

			statsTableData[i] = append(statsTableData[i],
				fmt.Sprintf("%s%%", humanize.FtoaWithDigits(stat.UserDetail.Share*100, 2)),
				fmt.Sprintf("%s", humanize.CommafWithDigits(stat.UserDetail.Reward/1e8, 3)))
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

	hashRate := int64(float64(opCount) / (float64(duration) / 1e9))
	return humanize.Comma(hashRate)
}
