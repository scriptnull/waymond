package direct

import (
	"context"
	"errors"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/connector"
	"github.com/scriptnull/waymond/internal/event"
)

var Type = connector.Type("direct")

type Connector struct {
	from string
	to   string
}

func (c *Connector) Type() connector.Type {
	return Type
}

func (c *Connector) Register(ctx context.Context) error {
	eventBus := ctx.Value("eventBus").(event.Bus)

	eventBus.Subscribe(c.from, func() {
		eventBus.Publish(c.to, []byte(""))
	})

	return nil
}

func ParseConfig(k *koanf.Koanf) (connector.Interface, error) {
	from := k.String("from")
	if from == "" {
		return nil, errors.New("expected non-empty value for 'from' in 'direct' connector")
	}

	to := k.String("to")
	if to == "" {
		return nil, errors.New("expected non-empty value for 'to' in 'direct' connector")
	}

	return &Connector{
		from,
		to,
	}, nil
}
