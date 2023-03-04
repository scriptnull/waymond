package docker

import (
	"errors"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "docker"

type Scaler struct {
	imageName string
	imageTag  string
	count     int
}

func (s *Scaler) Type() scaler.Type {
	return Type
}

func (s *Scaler) Register() error {
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
		imageName,
		imageTag,
		count,
	}, nil
}
