package main

import (
	"fmt"

	"github.com/buddyh/alexa-cli/internal/api"
	"github.com/spf13/cobra"
)

func newRoutineCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routine",
		Short: "Manage Alexa routines",
		Long:  `List and execute Alexa routines.`,
	}

	cmd.AddCommand(newRoutineListCmd(flags))
	cmd.AddCommand(newRoutineRunCmd(flags))

	return cmd
}

func newRoutineListCmd(flags *rootFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available routines",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			client, err := getClient()
			if err != nil {
				return err
			}

			routines, err := client.GetRoutines()
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.Data(routines)
			}

			if len(routines) == 0 {
				return out.Success("No routines found")
			}

			for _, r := range routines {
				fmt.Printf("  %s\n", r.Name)
			}
			return nil
		},
	}
}

func newRoutineRunCmd(flags *rootFlags) *cobra.Command {
	var device string

	cmd := &cobra.Command{
		Use:   "run <routine-name>",
		Short: "Execute a routine",
		Long: `Execute an Alexa routine by name.

Examples:
  alexacli routine run "Good Night"
  alexacli routine run "Morning Routine" -d Kitchen`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := getFormatter(flags)

			apiClient, err := getClient()
			if err != nil {
				return err
			}

			routineName := args[0]

			// Get a device to use for execution
			var dev *api.Device
			if device != "" {
				dev, err = findDevice(apiClient, device)
				if err != nil {
					return err
				}
			} else {
				// Use first available device
				devices, err := apiClient.GetDevices()
				if err != nil {
					return err
				}
				if len(devices) == 0 {
					return fmt.Errorf("no devices found")
				}
				dev = &devices[0]
			}

			if err := apiClient.ExecuteRoutine(dev, routineName); err != nil {
				return err
			}

			return out.Success(fmt.Sprintf("Executed routine: %s", routineName))
		},
	}

	cmd.Flags().StringVarP(&device, "device", "d", "", "Device to use for execution")

	return cmd
}
