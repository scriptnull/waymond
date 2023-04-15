package noop

import (
	"context"
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/log"
	"github.com/scriptnull/waymond/internal/scaler"
)

const Type scaler.Type = "noop"

type Scaler struct {
	id           string
	namespacedID string

	log log.Logger
}

func (s *Scaler) Type() scaler.Type {
	return Type
}

func (s *Scaler) Register(ctx context.Context) error {
	event.B.Subscribe(fmt.Sprintf("%s.input", s.namespacedID), func(data []byte) {
		s.log.Verbose("data = ", string(data))
	})
	return nil
}

func ParseConfig(k *koanf.Koanf) (scaler.Interface, error) {
	id := k.String("id")
	if id == "" {
		return nil, errors.New("expected non-empty value for 'id' in docker scaler")
	}

	s := &Scaler{
		id:           id,
		namespacedID: fmt.Sprintf("scaler.%s", id),
	}
	s.log = log.New(s.namespacedID)
	return s, nil
}
