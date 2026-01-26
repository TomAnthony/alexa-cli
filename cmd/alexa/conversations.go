package main

import (
	"fmt"

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

Examples:
  # List all conversations
  alexacli conversations

  # Output as JSON
  alexacli conversations --json

  # Use a conversation ID with askplus
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
				fmt.Printf("  ID:     %s\n\n", conv.ConversationID)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show debug output")

	return cmd
}
