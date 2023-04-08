package buildkite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/event"
	"github.com/scriptnull/waymond/internal/log"

	"github.com/scriptnull/waymond/internal/trigger"
)

const Type trigger.Type = "buildkite"

const tokenEnvName = "BUILDKITE_TOKEN"

type Trigger struct {
	log          log.Logger
	id           string
	namespacedID string

	req *http.Request
}

func (t *Trigger) Type() trigger.Type {
	return Type
}

func (t *Trigger) Register(ctx context.Context) error {
	buildkiteToken := os.Getenv(tokenEnvName)
	if len(buildkiteToken) == 0 {
		return fmt.Errorf("%s expects %s environment variable to be set while registering", t.namespacedID, tokenEnvName)
	}

	req, err := http.NewRequest("GET", "https://agent.buildkite.com/v3/metrics", nil)
	if err != nil {
		return fmt.Errorf("unable to construct request: %s", err)
	}
	req.Header.Add("User-Agent", "waymond-autoscaler")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", buildkiteToken))
	t.req = req

	event.B.Subscribe(fmt.Sprintf("%s.input", t.namespacedID), func(data []byte) {
		t.log.Debug("start of input event")
		_ = t.Do(data)
		t.log.Debug("end of input event")
	})

	return nil
}

func (t *Trigger) Do(_ []byte) error {
	resp, err := http.DefaultClient.Do(t.req)
	if err != nil {
		return fmt.Errorf("error in making buildkite api call: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected response status code: %d, but got: %d", http.StatusOK, resp.StatusCode)
	}

	// uncomment to check the full response body
	// by, _ := io.ReadAll(resp.Body)
	// t.log.Debug(string(by))

	var metrics struct {
		Jobs struct {
			Queues map[string]struct {
				Waiting int
			}
		} `json:"jobs"`
	}
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		return fmt.Errorf("unable to decode response body: %s", err)
	}

	for qName, q := range metrics.Jobs.Queues {
		t.log.Debugf("qName: %s, waitingSize: %d \n", qName, q.Waiting)
	}

	return nil
}

func ParseConfig(k *koanf.Koanf) (trigger.Interface, error) {
	id := k.String("id")
	if id == "" {
		return nil, errors.New("expected non-empty value for 'id' in buildkite trigger")
	}

	t := &Trigger{
		id:           id,
		namespacedID: fmt.Sprintf("trigger.%s", id),
	}
	t.log = log.New(t.namespacedID)

	return t, nil
}
