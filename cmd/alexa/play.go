package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newPlayCmd(flags *rootFlags) *cobra.Command {
	var device string
	var url string

	cmd := &cobra.Command{
		Use:   "play <file-or-url>",
		Short: "Play an audio file on an Alexa device",
		Long: `Play an MP3 audio file on an Alexa device using SSML.

The audio must be:
- MP3 format, 48kbps bitrate, 22050Hz sample rate
- Accessible via HTTPS with valid SSL certificate

If you provide a local file, it will be converted to the correct format.
You must provide a public HTTPS URL (use --url flag or host the file yourself).

Examples:
  # Play from a public URL
  alexacli play --url "https://example.com/audio.mp3" -d "Echo Show"

  # Convert local file (still needs hosting)
  alexacli play ~/audio.mp3 -d "Kitchen"`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			if url == "" && len(args) == 0 {
				return fmt.Errorf("provide either a URL (--url) or a local file path")
			}

			audioURL := url

			// If local file provided, convert it
			if len(args) > 0 && url == "" {
				localFile := args[0]
				if strings.HasPrefix(localFile, "~") {
					home, _ := os.UserHomeDir()
					localFile = filepath.Join(home, localFile[1:])
				}

				if _, err := os.Stat(localFile); os.IsNotExist(err) {
					return fmt.Errorf("file not found: %s", localFile)
				}

				// Convert to Alexa-compatible format
				outFile := filepath.Join(os.TempDir(), "alexa-audio-converted.mp3")
				convertCmd := exec.Command("ffmpeg", "-i", localFile, "-ar", "22050", "-ab", "48k", "-ac", "1", outFile, "-y")
				if output, err := convertCmd.CombinedOutput(); err != nil {
					return fmt.Errorf("failed to convert audio: %v\n%s", err, string(output))
				}

				fmt.Printf("Converted to: %s\n", outFile)
				fmt.Println("Note: You need to host this file on HTTPS and provide the URL with --url")
				return nil
			}

			if device == "" {
				return fmt.Errorf("device is required (use -d)")
			}

			client, err := getClient()
			if err != nil {
				return err
			}

			dev, err := findDevice(client, device)
			if err != nil {
				return err
			}

			// Send SSML with audio tag
			ssml := fmt.Sprintf(`<speak><audio src="%s"/></speak>`, audioURL)
			if err := client.SequenceCommand(dev, fmt.Sprintf("speak:'%s'", ssml)); err != nil {
				return err
			}

			return out.Success(fmt.Sprintf("Playing audio on %s", dev.AccountName))
		},
	}

	cmd.Flags().StringVarP(&device, "device", "d", "", "Device name or serial")
	cmd.Flags().StringVarP(&url, "url", "u", "", "HTTPS URL of the audio file")

	return cmd
}
