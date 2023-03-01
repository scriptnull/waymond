package requester

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
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
	return nil
}

func (b *Buildkite) AutoScaleRequest() error {
	fmt.Println("invoking buildkite AutoScaleRequest")
	req, err := http.NewRequest("GET", "https://agent.buildkite.com/v3/metrics", nil)
	if err != nil {
		return fmt.Errorf("debug new request error: %s", err)
	}
	req.Header.Add("User-Agent", "waymond-autoscaler")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", b.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("debug client do error: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status code is not okay: %d", resp.StatusCode)
	}

	// by, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(by))

	var agentMetrics struct {
		Organization struct {
			Slug string `json:"slug"`
		} `json:"organization"`
		Agents struct {
			Queues map[string]struct {
				Busy  int64 `json:"busy"`
				Idle  int64 `json:"idle"`
				Total int64 `json:"total"`
			} `json:"queues"`
		} `json:"agents"`
		Jobs struct {
			Queues map[string]struct {
				Scheduled int64 `json:"scheduled"`
				Running   int64 `json:"running"`
				Waiting   int64 `json:"waiting"`
			} `json:"queues"`
		} `json:"jobs"`
	}

	err = json.NewDecoder(resp.Body).Decode(&agentMetrics)
	if err != nil {
		return fmt.Errorf("unable to decode response body: %s", err)
	}

	fmt.Println("debug metrics response", agentMetrics)

	for queueName, queue := range agentMetrics.Jobs.Queues {
		if queue.Waiting > 0 {
			fmt.Printf("need %d agents for %s queue \n", queue.Waiting, queueName)
		}
	}
	return nil
}
