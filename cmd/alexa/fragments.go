package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newFragmentsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fragments <conversation-id>",
		Short: "Get fragments from an Alexa+ conversation",
		Long: `Retrieve all fragments from an Alexa+ conversation.

Fragments include both user messages and Alexa responses.
Use 'alexacli conversations' to find conversation IDs.

Examples:
  # Get fragments from a conversation
  alexacli fragments amzn1.conversation.xxx

  # Output as JSON for scripting
  alexacli fragments amzn1.conversation.xxx --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			client, err := getClient()
			if err != nil {
				return err
			}

			conversationID := args[0]
			client.SetConversationID(conversationID)

			resp, err := client.GetConversationFragments()
			if err != nil {
				return err
			}

			if flags.asJSON {
				// Build simplified output
				type fragmentOut struct {
					URI       string `json:"uri"`
					Timestamp string `json:"timestamp"`
					Purpose   string `json:"purpose"`
					Type      string `json:"type"`
					Text      string `json:"text"`
				}
				var frags []fragmentOut
				for _, f := range resp.Fragments {
					fo := fragmentOut{
						URI:       f.FragmentURI,
						Timestamp: f.Timestamp,
						Purpose:   f.Metadata.Purpose,
					}
					if f.Content != nil {
						fo.Type = f.Content.Type
						fo.Text = f.Content.GetText()
					}
					frags = append(frags, fo)
				}
				return out.Data(map[string]interface{}{
					"conversationId": resp.ConversationID,
					"fragments":      frags,
					"count":          len(frags),
				})
			}

			fmt.Printf("Conversation: %s\n", resp.ConversationID)
			fmt.Printf("Fragments: %d\n\n", len(resp.Fragments))

			for _, f := range resp.Fragments {
				text := ""
				if f.Content != nil {
					text = f.Content.GetText()
				}
				if text == "" {
					continue
				}

				// Format based on purpose
				switch f.Metadata.Purpose {
				case "USER":
					fmt.Printf("YOU: %s\n", text)
				case "AGENT":
					fmt.Printf("ALEXA: %s\n", truncateText(text, 200))
				default:
					fmt.Printf("[%s]: %s\n", f.Metadata.Purpose, truncateText(text, 100))
				}
				fmt.Println()
			}

			return nil
		},
	}

	return cmd
}

func truncateText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
