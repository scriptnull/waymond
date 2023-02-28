package requester

import (
	"errors"
	"os"
)

type Buildkite struct {
	token string

	AllowQueues  []string `json:"allow_queues"`
	RejectQueues []string `json:"reject_queues"`
}

func (b *Buildkite) Register() error {
	b.token = os.Getenv("BUILDKITE_TOKEN")
	if b.token == "" {
		return errors.New("BUILDKITE_TOKEN environment variable not set")
	}

	return nil
}
