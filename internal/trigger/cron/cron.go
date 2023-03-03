package cron

import (
	"errors"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/trigger"
)

const Type string = "cron"

type Trigger struct {
}

func ParseConfig(k *koanf.Koanf) (trigger.Interface, error) {
	expression := k.String("expression")
	if expression == "" {
		return nil, errors.New("expected non-empty value for 'expression' in cron trigger")
	}

	return &Trigger{}, nil
}
