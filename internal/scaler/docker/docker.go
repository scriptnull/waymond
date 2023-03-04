package docker

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "docker"

type Scaler struct {
	id        string
	imageName string
	imageTag  string
	count     int
}

func (s *Scaler) Type() scaler.Type {
	return Type
}

func (s *Scaler) Register(ctx context.Context) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	eventBus := ctx.Value("eventBus").(event.Bus)
	eventBus.Subscribe(fmt.Sprintf("scaler.%s", s.id), func() {
		fmt.Printf("scaler.%s.start \n", s.id)

		imageFullName := fmt.Sprintf("%s:%s", s.imageName, s.imageTag)

		containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
			Filters: filters.NewArgs(
				filters.KeyValuePair{
					Key:   "status",
					Value: "running",
				},
				filters.KeyValuePair{
					Key:   "ancestor",
					Value: imageFullName,
				},
			),
		})
		if err != nil {
			fmt.Println("error listing containers", err)
			return
		}

		currentCount := len(containers)
		if currentCount < s.count {
			fmt.Printf("scaler.%s current count: %d, desired count: %d \n", s.id, currentCount, s.count)
			remainingCount := s.count - currentCount

			for i := 0; i < remainingCount; i++ {
				c, err := cli.ContainerCreate(ctx, &container.Config{Image: imageFullName}, nil, nil, nil, "")
				if err != nil {
					fmt.Printf("scaler.%s.error: %s \n", s.id, err)
					return
				}
				fmt.Printf("scaler.%s.container.created: %s \n", s.id, c.ID)

				err = cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{})
				if err != nil {
					fmt.Printf("scaler.%s.error: %s \n", s.id, err)
					return
				}
				fmt.Printf("scaler.%s.container.started: %s \n", s.id, c.ID)
			}
		}

		fmt.Printf("scaler.%s.done \n", s.id)
	})
	return nil
}

func ParseConfig(k *koanf.Koanf) (scaler.Interface, error) {
	imageName := k.String("image_name")
	if imageName == "" {
		return nil, errors.New("expected non-empty value for 'image_name' in cron trigger")
	}

	imageTag := k.String("image_tag")
	if imageTag == "" {
		return nil, errors.New("expected non-empty value for 'image_tag' in cron trigger")
	}

	count := k.Int("count")

	return &Scaler{
		k.String("id"),
		imageName,
		imageTag,
		count,
	}, nil
}
