package transform

import (
	"errors"
	"fmt"

	"github.com/knadh/koanf/v2"
)

var (
	ErrUnknownTransformer   = errors.New("unknown transformer")
	ErrInvalidConfiguration = errors.New("invalid configuration")
)

type Transformer interface {
	Transform([]byte) ([]byte, error)
}

func ParseConfig(k *koanf.Koanf) (Transformer, error) {
	switch k.String("transform.method") {
	case "go_template":
		return newGoTemplate(k)
	default:
		return nil, ErrUnknownTransformer
	}
}

func newGoTemplate(k *koanf.Koanf) (*goTemplate, error) {
	templateFieldPath := "transform.template"
	template := k.String(templateFieldPath)
	if template == "" {
		return nil, fmt.Errorf("%w: missing %s", ErrInvalidConfiguration, templateFieldPath)
	}

	return &goTemplate{
		template: template,
	}, nil
}

type goTemplate struct {
	template string
}

func (g *goTemplate) Transform(inputData []byte) ([]byte, error) {
	return nil, nil
}
