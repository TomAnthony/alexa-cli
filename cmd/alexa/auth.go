package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/buddyh/alexa-cli/internal/api"
	"github.com/buddyh/alexa-cli/internal/config"
	"github.com/spf13/cobra"
)

func newAuthCmd(flags *rootFlags) *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "auth [refresh-token]",
		Short: "Configure authentication",
		Long: `Configure the Alexa CLI with your refresh token.

To obtain a refresh token, use the alexa-cookie-cli tool:
  npx alexa-cookie-cli

This will open a browser for you to log into Amazon, then provide
the refresh token to paste here.

Alternatively, set the ALEXA_REFRESH_TOKEN environment variable.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			var token string
			if len(args) > 0 {
				token = args[0]
			} else {
				// Prompt for token
				fmt.Print("Enter refresh token: ")
				reader := bufio.NewReader(os.Stdin)
				input, err := reader.ReadString('\n')
				if err != nil {
					return fmt.Errorf("failed to read input: %w", err)
				}
				token = strings.TrimSpace(input)
			}

			if token == "" {
				return fmt.Errorf("refresh token is required")
			}

			// Validate the token by trying to authenticate
			client, err := api.NewClient(token, domain)
			if err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Try to get devices to verify it works
			devices, err := client.GetDevices()
			if err != nil {
				return fmt.Errorf("failed to verify token: %w", err)
			}

			// Save configuration
			cfg := &config.Config{
				RefreshToken: token,
				AmazonDomain: domain,
			}

			if err := config.Save(cfg); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			path, _ := config.Path()
			return out.Success(fmt.Sprintf("Authenticated successfully. Found %d devices. Config saved to %s", len(devices), path))
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "amazon.com", "Amazon domain (amazon.com, amazon.de, amazon.co.uk, etc.)")

	return cmd
}
