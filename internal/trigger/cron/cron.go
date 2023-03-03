package cron

import (
	"errors"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/trigger"
)

const Type trigger.Type = "cron"

type Trigger struct {
}

func (t *Trigger) Type() trigger.Type {
	return Type
}

func (t *Trigger) Register() error {
	return nil
}

func ParseConfig(k *koanf.Koanf) (trigger.Interface, error) {
	expression := k.String("expression")
	if expression == "" {
		return nil, errors.New("expected non-empty value for 'expression' in cron trigger")
	}

	return &Trigger{}, nil
}
