package cmd

import (
	"context"
	"encoding/base64"
	"github.com/digitalocean/godo"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/digitalocean"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/spf13/cobra"
)

// CreateCmd holds the cmd flags
type CreateCmd struct{}

// NewCreateCmd defines a command
func NewCreateCmd() *cobra.Command {
	cmd := &CreateCmd{}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv()
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return createCmd
}

// Run runs the command logic
func (cmd *CreateCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	req, err := buildInstance(options)
	if err != nil {
		return err
	}

	return digitalocean.NewDigitalOcean(options.Token).Create(ctx, req)
}

func GetInjectKeypairScript(dir string) (string, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(dir)
	if err != nil {
		return "", err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return "", err
	}

	resultScript := `#!/bin/sh
useradd devpod -d /home/devpod
mkdir -p /home/devpod
if grep -q sudo /etc/groups; then
	usermod -aG sudo devpod
elif grep -q wheel /etc/groups; then
	usermod -aG wheel devpod
fi
echo "devpod ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-devpod
mkdir -p /home/devpod/.ssh
echo "` + string(publicKey) + `" >> /home/devpod/.ssh/authorized_keys
chmod 0700 /home/devpod/.ssh
chmod 0600 /home/devpod/.ssh/authorized_keys
chown -R devpod:devpod /home/devpod`

	return resultScript, nil
}

func buildInstance(options *options.Options) (*godo.DropletCreateRequest, error) {
	// generate ssh keys
	userData, err := GetInjectKeypairScript(options.MachineFolder)
	if err != nil {
		return nil, err
	}

	// generate instance object
	instance := &godo.DropletCreateRequest{
		Name:   options.MachineID,
		Region: options.Region,
		Size:   options.MachineType,
		Image: godo.DropletCreateImage{
			Slug: options.DiskImage,
		},
		UserData: userData,
		Tags:     []string{"devpod"},
	}

	return instance, nil
}
