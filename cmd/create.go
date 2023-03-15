package cmd

import (
	"context"
	"encoding/base64"
	"github.com/digitalocean/godo"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/digitalocean"
	"github.com/loft-sh/devpod-provider-digitalocean/pkg/options"
	"github.com/loft-sh/devpod/pkg/log"
	"github.com/loft-sh/devpod/pkg/ssh"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strconv"
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

	diskSize, err := strconv.Atoi(options.DiskSize)
	if err != nil {
		return errors.Wrap(err, "parse disk size")
	}

	return digitalocean.NewDigitalOcean(options.Token).Create(ctx, req, diskSize)
}

func GetInjectKeypairScript(machineFolder, machineID string) (string, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(machineFolder)
	if err != nil {
		return "", err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return "", err
	}

	resultScript := `#!/bin/sh

# Mount volume to home
mkdir -p /home/devpod
mount -o discard,defaults,noatime /dev/disk/by-id/scsi-0DO_Volume_` + machineID + ` /home/devpod

# Move docker data dir
service docker stop
cat > /etc/docker/daemon.json << EOF
{
  "data-root": "/home/devpod/.docker-daemon",
  "live-restore": true
}
EOF
# Make sure we only copy if volumes isn't initialized
if [ ! -d "/home/devpod/.docker-daemon" ]; then
  mkdir -p /home/devpod/.docker-daemon
  rsync -aP /var/lib/docker/ /home/devpod/.docker-daemon
fi
service docker start

# Create DevPod user and configure ssh
useradd devpod -d /home/devpod
if grep -q sudo /etc/groups; then
	usermod -aG sudo devpod
elif grep -q wheel /etc/groups; then
	usermod -aG wheel devpod
fi
echo "devpod ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-devpod
mkdir -p /home/devpod/.ssh
echo '` + string(publicKey) + `' > /home/devpod/.ssh/authorized_keys
chmod 0700 /home/devpod/.ssh
chmod 0600 /home/devpod/.ssh/authorized_keys
chown devpod:devpod /home/devpod
chown -R devpod:devpod /home/devpod/.ssh

# Make sure we don't get limited
ufw allow 22/tcp || true
`

	return resultScript, nil
}

func buildInstance(options *options.Options) (*godo.DropletCreateRequest, error) {
	// generate ssh keys
	userData, err := GetInjectKeypairScript(options.MachineFolder, options.MachineID)
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
