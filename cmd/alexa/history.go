package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func newHistoryCmd(flags *rootFlags) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent Alexa activity history",
		Long: `Display recent voice commands and Alexa's responses.

Shows what you said (or what was sent via textcommand) and what Alexa
responded with.

Examples:
  alexacli history
  alexacli history --limit 5
  alexacli history --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			client, err := getClient()
			if err != nil {
				return err
			}

			// Get devices first to ensure we have customer ID
			_, err = client.GetDevices()
			if err != nil {
				return err
			}

			// Get last 24 hours of history
			endTime := time.Now().UnixMilli()
			startTime := endTime - (24 * 60 * 60 * 1000) // 24 hours ago

			records, err := client.GetCustomerHistoryRecords(startTime, endTime)
			if err != nil {
				return err
			}

			if flags.asJSON {
				if limit > 0 && len(records) > limit {
					records = records[:limit]
				}
				return out.Data(records)
			}

			if len(records) == 0 {
				return out.Success("No recent activity found")
			}

			count := 0
			for _, r := range records {
				if limit > 0 && count >= limit {
					break
				}

				t := time.UnixMilli(r.Timestamp)
				fmt.Printf("[%s] Device: %s\n", t.Format("15:04:05"), r.Device)
				if r.CustomerUtterance != "" {
					fmt.Printf("  You: %s\n", r.CustomerUtterance)
				}
				if r.AlexaResponse != "" {
					fmt.Printf("  Alexa: %s\n", r.AlexaResponse)
				}
				fmt.Println()
				count++
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Maximum number of records to show")

	return cmd
}
