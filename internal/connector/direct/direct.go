package direct

import (
	"context"
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/connector"
	"github.com/scriptnull/waymond/internal/connector/transform"
	"github.com/scriptnull/waymond/internal/event"
)

var Type = connector.Type("direct")

type Connector struct {
	id   string
	from string
	to   string

	transformer transform.Transformer
}

func (c *Connector) Type() connector.Type {
	return Type
}

func (c *Connector) Register(ctx context.Context) error {
	event.B.Subscribe(fmt.Sprintf("%s.output", c.from), func(inputData []byte) {
		var err error
		outputData := inputData
		if c.transformer != nil {
			outputData, err = c.transformer.Transform(inputData)
			if err != nil {
				event.B.Publish(fmt.Sprintf("connector.%s.error", c.id), []byte(err.Error()))
			}
		}
		event.B.Publish(fmt.Sprintf("%s.input", c.to), outputData)
	})

	return nil
}

func ParseConfig(k *koanf.Koanf) (connector.Interface, error) {
	id := k.String("id")
	if id == "" {
		return nil, errors.New("expected non-empty value for 'id' in 'direct' connector")
	}

	from := k.String("from")
	if from == "" {
		return nil, errors.New("expected non-empty value for 'from' in 'direct' connector")
	}

	to := k.String("to")
	if to == "" {
		return nil, errors.New("expected non-empty value for 'to' in 'direct' connector")
	}

	c := &Connector{
		id:   id,
		from: from,
		to:   to,
	}

	if k.Exists("transform") {
		transformer, err := transform.ParseConfig(k)
		if err != nil {
			return nil, err
		}
		c.transformer = transformer
	}

	return c, nil
}
