package cmd

import (
	"context"
	"fmt"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/digitalocean"
	"os"

	"github.com/loft-sh/devpod-provider-digitalocean/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// CommandCmd holds the cmd flags
type CommandCmd struct{}

// NewCommandCmd defines a command
func NewCommandCmd() *cobra.Command {
	cmd := &CommandCmd{}
	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "Run a command on the instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(false)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return commandCmd
}

// Run runs the command logic
func (cmd *CommandCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	command := os.Getenv("COMMAND")
	if command == "" {
		return fmt.Errorf("command environment variable is missing")
	}

	// get private key
	privateKey, err := ssh.GetPrivateKeyRawBase(options.MachineFolder)
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	// create client
	droplet, err := digitalocean.NewDigitalOcean(options.Token).GetByName(ctx, options.MachineID)
	if err != nil {
		return err
	} else if droplet == nil {
		return fmt.Errorf("droplet not found")
	}

	// get external ip
	if droplet.Networks == nil {
		return fmt.Errorf("couldn't find public ip address")
	}

	// loop over addresses
	externalIP := ""
	for _, network := range droplet.Networks.V4 {
		if network.Type == "public" && network.IPAddress != "" {
			externalIP = network.IPAddress
			break
		}
	}
	if externalIP == "" {
		return fmt.Errorf("couldn't find a public ip address")
	}

	// dial external address
	sshClient, err := ssh.NewSSHClient("devpod", externalIP+":22", privateKey)
	if err != nil {
		return errors.Wrap(err, "create ssh client")
	}
	defer sshClient.Close()

	// run command
	return ssh.Run(context.Background(), sshClient, command, os.Stdin, os.Stdout, os.Stderr)
}
