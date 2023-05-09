package awsec2asg

import (
	"context"
	"errors"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/log"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "aws_ec2_asg"

type Scaler struct {
	id           string
	namespacedID string
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

	event.B.Subscribe(s.namespacedID, func(data []byte) {
		s.log.Verbose("start")

		s.log.Debugf("data: %+v\n", string(data))

		s.log.Verbose("end")
	})
	return nil
}

func ParseConfig(k *koanf.Koanf) (scaler.Interface, error) {
	id := k.String("id")
	if id == "" {
		return nil, errors.New("expected non-empty value for 'id' in aws_ec2_asg scaler")
	}

	s := &Scaler{
		id:           id,
		namespacedID: fmt.Sprintf("scaler.%s", id),
	}
	s.log = log.New(s.namespacedID)
	return s, nil
}
