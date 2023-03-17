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

// DeleteCmd holds the cmd flags
type DeleteCmd struct{}

// NewDeleteCmd defines a command
func NewDeleteCmd() *cobra.Command {
	cmd := &DeleteCmd{}
	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(false)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return deleteCmd
}

// Run runs the command logic
func (cmd *DeleteCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	digitalOceanClient := digitalocean.NewDigitalOcean(options.Token)
	err := digitalOceanClient.Delete(ctx, options.MachineID)
	if err != nil {
		return err
	}

	// wait until deleted
	for {
		status, err := digitalOceanClient.Status(ctx, options.MachineID)
		if err != nil {
			log.Errorf("Error retrieving droplet status: %v", err)
			break
		} else if status == client.StatusNotFound {
			break
		}

		// make sure we don't spam
		time.Sleep(time.Second)
	}

	return nil
}
