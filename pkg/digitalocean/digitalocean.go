package digitalocean

import (
	"context"
	"fmt"
	"github.com/digitalocean/godo"
	"github.com/loft-sh/devpod/pkg/client"
)

func NewDigitalOcean(token string) *DigitalOcean {
	return &DigitalOcean{
		client: godo.NewFromToken(token),
	}
}

type DigitalOcean struct {
	client *godo.Client
}

func (d *DigitalOcean) Create(ctx context.Context, req *godo.DropletCreateRequest) error {
	_, _, err := d.client.Droplets.Create(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (d *DigitalOcean) Status(ctx context.Context, name string) (client.Status, error) {
	droplet, err := d.GetByName(ctx, name)
	if err != nil {
		return client.StatusNotFound, err
	} else if droplet == nil {
		// TODO: Check for snapshot
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
