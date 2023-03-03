package cron

import (
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"
	"github.com/scriptnull/waymond/internal/trigger"
)

const Type trigger.Type = "cron"

type Trigger struct {
	cronExpr string
}

func (t *Trigger) Type() trigger.Type {
	return Type
}

func (t *Trigger) Register() error {
	c := cron.New()
	c.AddFunc(t.cronExpr, func() {
		t.Do()
	})
	c.Start()
	return nil
}

func (t *Trigger) Do() error {
	fmt.Println("Cron trigger Do func called")
	return nil
}

func ParseConfig(k *koanf.Koanf) (trigger.Interface, error) {
	expression := k.String("expression")
	if expression == "" {
		return nil, errors.New("expected non-empty value for 'expression' in cron trigger")
	}

	return &Trigger{
		cronExpr: expression,
	}, nil
}
