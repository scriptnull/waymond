package requester

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/scriptnull/waymond/schedule"
)

type Instance struct {
	Type     string            `mapstructure:"type"`
	Schedule schedule.Schedule `mapstructure:"schedule"`
	Config   map[string]any    `mapstructure:"config"`
}

var ErrUnknownType = errors.New("unknown requester type")

func (ins *Instance) Register() error {
	type Requester interface {
		AutoScaleRequest()
	}

	var requester Requester
	switch ins.Type {
	case "buildkite":
		var bkRequester Buildkite
		err := mapstructure.Decode(ins.Config, &bkRequester)
		if err != nil {
			return fmt.Errorf("unable to decode buildkite config, %s", err)
		}
		err = bkRequester.Register()
		if err != nil {
			return fmt.Errorf("unable to register buildkite requestor: %s", err)
		}

		requester = &bkRequester
	default:
		return ErrUnknownType
	}

	return ins.Schedule.Register(requester.AutoScaleRequest)
}
