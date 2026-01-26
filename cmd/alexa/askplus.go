package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newAskPlusCmd(flags *rootFlags) *cobra.Command {
	var timeout int
	var conversationID string
	var verbose bool

	cmd := &cobra.Command{
		Use:   "askplus <question>",
		Short: "Ask Alexa+ (LLM) a question and get an AI response",
		Long: `Have a two-way conversation with Alexa+ (LLM-powered assistant).

This command uses the Alexa+ LLM backend, which provides:
- Conversational AI responses with markdown formatting
- Multi-turn conversations with persistent context
- Source citations when applicable
- Complex reasoning, math, and creative tasks

REQUIRES: A conversation ID from 'alexacli conversations'. The conversation
ID persists and allows ongoing chat sessions.

Examples:
  # Ask Alexa+ a question
  alexacli askplus -c "amzn1.conversation.xxx" "What is the capital of France?"

  # Complex reasoning
  alexacli askplus -c "amzn1.conversation.xxx" "If I have 12 cookies and give away a third, how many left?"

  # Creative tasks
  alexacli askplus -c "amzn1.conversation.xxx" "Write a haiku about debugging"

  # With longer timeout
  alexacli askplus -c "amzn1.conversation.xxx" -t 30 "Explain quantum computing"`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			client, err := getClient()
			if err != nil {
				return err
			}

			// Set conversation ID if provided
			if conversationID != "" {
				client.SetConversationID(conversationID)
			}

			// Enable verbose mode
			if verbose {
				client.SetVerbose(true)
			}

			question := strings.Join(args, " ")
			timeoutDuration := time.Duration(timeout) * time.Second

			response, err := client.AskPlus(question, timeoutDuration)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.Data(map[string]string{
					"question": question,
					"response": response,
					"type":     "alexa_plus",
				})
			}

			fmt.Println(response)
			return nil
		},
	}

	cmd.Flags().IntVarP(&timeout, "timeout", "t", 15, "Timeout in seconds to wait for response")
	cmd.Flags().StringVarP(&conversationID, "conversation", "c", "", "Existing conversation ID for multi-turn")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show debug output")

	return cmd
}
