package transform

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/knadh/koanf/v2"
)

var (
	ErrUnknownTransformer   = errors.New("unknown transformer")
	ErrInvalidConfiguration = errors.New("invalid configuration")
	ErrNotAJSONInput        = errors.New("not a json input")
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
	templateString := k.String(templateFieldPath)
	if templateString == "" {
		return nil, fmt.Errorf("%w: missing %s", ErrInvalidConfiguration, templateFieldPath)
	}

	return &goTemplate{
		template: templateString,
	}, nil
}

type goTemplate struct {
	template string
}

func (g *goTemplate) Transform(inputData []byte) ([]byte, error) {
	var input interface{}
	err := json.Unmarshal(inputData, &input)
	if err != nil {
		return nil, ErrNotAJSONInput
	}

	// TODO: maybe use a global instance of template while tranforming
	templ := template.Must(template.New("transform").Parse(g.template))
	buf := bytes.NewBuffer([]byte(""))
	templ.Execute(buf, input)
	return buf.Bytes(), nil
}
