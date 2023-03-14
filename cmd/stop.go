package cmd

import (
	"context"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/digitalocean"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/options"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/spf13/cobra"
	"time"
)

// StopCmd holds the cmd flags
type StopCmd struct{}

// NewStopCmd defines a command
func NewStopCmd() *cobra.Command {
	cmd := &StopCmd{}
	stopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv()
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return stopCmd
}

// Run runs the command logic
func (cmd *StopCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	digitalOceanClient := digitalocean.NewDigitalOcean(options.Token)
	err := digitalOceanClient.Stop(ctx, options.MachineID)
	if err != nil {
		return err
	}

	// wait until stopped
	for {
		status, err := digitalOceanClient.Status(ctx, options.MachineID)
		if err != nil {
			log.Errorf("Error retrieving droplet status: %v", err)
			break
		} else if status == client.StatusStopped {
			break
		}

		// make sure we don't spam
		time.Sleep(time.Second)
	}

	return nil
}
