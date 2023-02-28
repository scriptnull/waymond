package requester

import (
	"errors"
	"fmt"
	"os"
)

type Buildkite struct {
	token string

	AllowQueues  []string `mapstructure:"allow_queues"`
	RejectQueues []string `mapstructure:"reject_queues"`
}

func (b *Buildkite) Register() error {
	b.token = os.Getenv("BUILDKITE_TOKEN")
	if b.token == "" {
		return errors.New("BUILDKITE_TOKEN environment variable not set")
	}

	fmt.Println("buildkite config:", b)

	return nil
}

func (b *Buildkite) AutoScaleRequest() {
	fmt.Println("TODO: make request to buildkite to determine whether to auto-scale or not")
}
