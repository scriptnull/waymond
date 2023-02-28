package schedule

import (
	"errors"
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/robfig/cron/v3"
)

var ErrUnknownType = errors.New("unknown schedule type")

var CronScheduler *cron.Cron

type Schedule struct {
	Type   string         `mapstructure:"type"`
	Config map[string]any `mapstructure:"config"`
}

func (s *Schedule) Register(autoScaleRequest func()) error {
	fmt.Println("debug schedule", s)
	switch s.Type {
	case "cron":
		var c CronConfig
		err := mapstructure.Decode(s.Config, &c)
		if err != nil {
			return fmt.Errorf("unable to decode cron config: %s", err)
		}

		if CronScheduler == nil {
			CronScheduler = cron.New()
		}
		fmt.Println("debug cron config", c)
		_, err = CronScheduler.AddFunc(c.Expression, autoScaleRequest)
		return err
	default:
		return ErrUnknownType
	}
}

type CronConfig struct {
	Expression string `mapstrucuture:"expression"`
}
