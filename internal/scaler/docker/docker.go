package docker

import (
	"context"
	"errors"
	"fmt"

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
	eventBus := ctx.Value("eventBus").(event.Bus)
	eventBus.Subscribe(fmt.Sprintf("scaler.%s", s.id), func() {
		fmt.Println("event received inside docker scaler. this will perform an autoscale")
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
