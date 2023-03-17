package digitalocean

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/loft-sh/devpod/pkg/client"
	"github.com/pkg/errors"
	"time"
)

func NewDigitalOcean(token string) *DigitalOcean {
	return &DigitalOcean{
		client: godo.NewFromToken(token),
	}
}

type DigitalOcean struct {
	client *godo.Client
}

func (d *DigitalOcean) Init(ctx context.Context) error {
	_, _, err := d.client.Droplets.List(ctx, &godo.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "list droplets")
	}

	return nil
}

func (d *DigitalOcean) Create(ctx context.Context, req *godo.DropletCreateRequest, diskSize int) error {
	// create volume
	volume, err := d.volumeByName(ctx, req.Name)
	if err != nil {
		return err
	} else if volume == nil {
		volume, _, err = d.client.Storage.CreateVolume(ctx, &godo.VolumeCreateRequest{
			Region:          req.Region,
			Name:            req.Name,
			SizeGigaBytes:   int64(diskSize),
			FilesystemType:  "ext4",
			FilesystemLabel: "DevPod Data",
			Tags:            []string{"devpod"},
		})
		if err != nil {
			return errors.Wrap(err, "create volume")
		}
	}

	// create droplet
	req.Volumes = append(req.Volumes, godo.DropletCreateVolume{
		ID: volume.ID,
	})
	_, _, err = d.client.Droplets.Create(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (d *DigitalOcean) volumeByName(ctx context.Context, name string) (*godo.Volume, error) {
	volumes, _, err := d.client.Storage.ListVolumes(ctx, &godo.ListVolumeParams{Name: name})
	if err != nil {
		return nil, err
	} else if len(volumes) > 1 {
		return nil, fmt.Errorf("multiple volumes with name %s found", name)
	} else if len(volumes) == 0 {
		return nil, nil
	}

	return &volumes[0], nil
}

func (d *DigitalOcean) Stop(ctx context.Context, name string) error {
	droplet, err := d.GetByName(ctx, name)
	if err != nil {
		return err
	} else if droplet == nil {
		return nil
	}

	_, err = d.client.Droplets.Delete(ctx, droplet.ID)
	if err != nil {
		return err
	}

	return nil
}

func (d *DigitalOcean) Status(ctx context.Context, name string) (client.Status, error) {
	// get droplet
	droplet, err := d.GetByName(ctx, name)
	if err != nil {
		return client.StatusNotFound, err
	} else if droplet == nil {
		// check if volume exists
		volume, err := d.volumeByName(ctx, name)
		if err != nil {
			return client.StatusNotFound, err
		} else if volume != nil {
			return client.StatusStopped, nil
		}

		return client.StatusNotFound, nil
	}

	// is busy?
	if droplet.Status != "active" {
		return client.StatusBusy, nil
	}

	return client.StatusRunning, nil
}

func (d *DigitalOcean) GetByName(ctx context.Context, name string) (*godo.Droplet, error) {
	droplets, _, err := d.client.Droplets.ListByName(ctx, name, &godo.ListOptions{})
	if err != nil {
		return nil, err
	} else if len(droplets) > 1 {
		return nil, fmt.Errorf("multiple droplets with name %s found", name)
	} else if len(droplets) == 0 {
		return nil, nil
	}

	return &droplets[0], nil
}

func (d *DigitalOcean) Delete(ctx context.Context, name string) error {
	// delete volume
	volume, err := d.volumeByName(ctx, name)
	if err != nil {
		return err
	} else if volume != nil {
		// detach volume
		for _, dropletID := range volume.DropletIDs {
			_, _, err = d.client.StorageActions.DetachByDropletID(ctx, volume.ID, dropletID)
			if err != nil {
				return errors.Wrap(err, "detach volume")
			}
		}

		// wait until volume is detached
		for len(volume.DropletIDs) > 0 {
			time.Sleep(time.Second)

			// re-get volume
			volume, err = d.volumeByName(ctx, name)
			if err != nil {
				return err
			} else if volume == nil {
				break
			}
		}

		// delete volume
		if volume != nil {
			_, err = d.client.Storage.DeleteVolume(ctx, volume.ID)
			if err != nil {
				return errors.Wrap(err, "delete volume")
			}
		}
	}

	droplet, err := d.GetByName(ctx, name)
	if err != nil {
		return err
	} else if droplet == nil {
		return nil
	}

	_, err = d.client.Droplets.Delete(ctx, droplet.ID)
	if err != nil {
		return err
	}

	return nil
}
