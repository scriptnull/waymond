package requester

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type Instance struct {
	Type   string         `mapstructure:"type"`
	Config map[string]any `mapstructure:",remain"`
}

var ErrUnknownType = errors.New("unknown requester type")

func (ins *Instance) Register() error {
	switch ins.Type {
	case "buildkite":
		var bkRequester Buildkite
		err := mapstructure.Decode(ins.Config, &bkRequester)
		if err != nil {
			return fmt.Errorf("unable to decode buildkite config, %s", err)
		}
		return bkRequester.Register()
	default:
		return ErrUnknownType
	}
}
