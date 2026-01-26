package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newConversationsCmd(flags *rootFlags) *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "conversations",
		Short: "List all Alexa+ conversations with device names",
		Long: `List all Alexa+ (LLM) conversation IDs and their associated devices.

This command retrieves all active Alexa+ conversations from the AVS API.
Each conversation ID is linked to a specific device and can be used with
the 'askplus' command for continued conversations.

TIP: Use 'alexacli askplus -d "Device Name"' to auto-select the most recent
conversation for a device, instead of manually copying conversation IDs.

Examples:
  # List all conversations
  alexacli conversations

  # Output as JSON
  alexacli conversations --json

  # Use askplus with device name (easier)
  alexacli askplus -d "Echo Show" "Hello"

  # Or use a specific conversation ID
  alexacli askplus -c "amzn1.conversation.xxx" "Hello"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			client, err := getClient()
			if err != nil {
				return err
			}

			if verbose {
				client.SetVerbose(true)
			}

			conversations, err := client.GetConversations()
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.Data(map[string]interface{}{
					"conversations": conversations,
					"count":         len(conversations),
				})
			}

			if len(conversations) == 0 {
				fmt.Println("No Alexa+ conversations found.")
				return nil
			}

			fmt.Printf("Found %d Alexa+ conversation(s):\n\n", len(conversations))
			for _, conv := range conversations {
				fmt.Printf("  Device: %s\n", conv.DeviceName)
				fmt.Printf("  ID:     %s\n", conv.ConversationID)
				if conv.LastTurnTime != "" {
					fmt.Printf("  Last:   %s\n", formatTime(conv.LastTurnTime))
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show debug output")

	return cmd
}

// formatTime converts an ISO timestamp to a human-readable format
func formatTime(isoTime string) string {
	// Try parsing ISO 8601 format
	t, err := time.Parse(time.RFC3339, isoTime)
	if err != nil {
		// Try without timezone
		t, err = time.Parse("2006-01-02T15:04:05.000Z", isoTime)
		if err != nil {
			// Try with milliseconds
			if strings.Contains(isoTime, ".") {
				t, err = time.Parse("2006-01-02T15:04:05.000Z07:00", isoTime)
			}
			if err != nil {
				return isoTime // Return as-is if parsing fails
			}
		}
	}
	return t.Local().Format("2006-01-02 15:04:05")
}
