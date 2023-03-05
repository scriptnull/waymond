package cron

import (
	"context"
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/trigger"
)

const Type trigger.Type = "cron"

type Trigger struct {
	id       string
	cronExpr string
}

func (t *Trigger) Type() trigger.Type {
	return Type
}

func (t *Trigger) Register(ctx context.Context) error {
	c := cron.New()
	_, err := c.AddFunc(t.cronExpr, func() {
		event.B.Publish(fmt.Sprintf("trigger.%s", t.id), []byte{})
	})
	if err != nil {
		return err
	}
	c.Start()
	return nil
}

func ParseConfig(k *koanf.Koanf) (trigger.Interface, error) {
	expression := k.String("expression")
	if expression == "" {
		return nil, errors.New("expected non-empty value for 'expression' in cron trigger")
	}

	return &Trigger{
		id:       k.String("id"),
		cronExpr: expression,
	}, nil
}
