package cmd

import (
	"context"
	"fmt"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/digitalocean"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/spf13/cobra"
	"os"
)

// StatusCmd holds the cmd flags
type StatusCmd struct{}

// NewStatusCmd defines a command
func NewStatusCmd() *cobra.Command {
	cmd := &StatusCmd{}
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Retrieve the status of an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv()
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return statusCmd
}

// Run runs the command logic
func (cmd *StatusCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	status, err := digitalocean.NewDigitalOcean(options.Token).Status(ctx, options.MachineID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, status)
	return err
}
