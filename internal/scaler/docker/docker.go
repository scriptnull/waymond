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
	"github.com/scriptnull/waymond/internal/log"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "docker"

type Scaler struct {
	id           string
	namespacedID string
	imageName    string
	imageTag     string
	count        int
	log          log.Logger
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

	event.B.Subscribe(s.namespacedID, func() {
		s.log.Verbose("start")

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
			s.log.Errorf("error listing containers", err)
			return
		}

		currentCount := len(containers)
		if currentCount < s.count {
			s.log.Debugf("current count: %d, desired count: %d \n", currentCount, s.count)
			remainingCount := s.count - currentCount
			s.log.Debugf("scaling up by creating %d container(s) \n", remainingCount)

			for i := 0; i < remainingCount; i++ {
				c, err := cli.ContainerCreate(ctx, &container.Config{Image: imageFullName}, nil, nil, nil, "")
				if err != nil {
					s.log.Errorf("error: %s \n", err)
					return
				}
				s.log.Debugf("container.created: %s \n", c.ID)

				err = cli.ContainerStart(ctx, c.ID, types.ContainerStartOptions{})
				if err != nil {
					s.log.Errorf("error: %s \n", err)
					return
				}
				s.log.Debugf("container.started: %s \n", c.ID)
			}
		} else if currentCount > s.count {
			s.log.Debugf("current count: %d, desired count: %d \n", currentCount, s.count)
			deletionCount := currentCount - s.count
			s.log.Debugf("scaling down by removing %d container(s) \n", deletionCount)

			for i := 0; i < deletionCount; i++ {
				err = cli.ContainerRemove(ctx, containers[i].ID, types.ContainerRemoveOptions{Force: true})
				if err != nil {
					s.log.Errorf("error: %s \n", err)
					return
				}
				s.log.Debugf("container.removed: %s \n", containers[i].ID)
			}
		}

		s.log.Verbose("end")
	})
	return nil
}

func ParseConfig(k *koanf.Koanf) (scaler.Interface, error) {
	id := k.String("id")
	if id == "" {
		return nil, errors.New("expected non-empty value for 'id' in docker scaler")
	}

	imageName := k.String("image_name")
	if imageName == "" {
		return nil, errors.New("expected non-empty value for 'image_name' in docker scaler")
	}

	imageTag := k.String("image_tag")
	if imageTag == "" {
		return nil, errors.New("expected non-empty value for 'image_tag' in docker scaler")
	}

	count := k.Int("count")

	s := &Scaler{
		id:           id,
		namespacedID: fmt.Sprintf("scaler.%s", id),
		imageName:    imageName,
		imageTag:     imageTag,
		count:        count,
	}
	s.log = log.New(s.namespacedID)
	return s, nil
}
