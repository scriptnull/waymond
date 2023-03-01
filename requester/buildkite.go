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

func (b *Buildkite) AutoScaleRequest() {
	fmt.Println("invoking buildkite AutoScaleRequest")
	req, err := http.NewRequest("GET", "https://agent.buildkite.com/v3/metrics", nil)
	if err != nil {
		fmt.Println("debug new request error:", err)
	}
	req.Header.Add("User-Agent", "waymond-autoscaler")
	req.Header.Add("Authorization", fmt.Sprintf("Token %s", b.token))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("debug client do error:", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Println("status code is not okay: ", resp.StatusCode)
		return
	}

	// by, _ := io.ReadAll(resp.Body)
	// fmt.Println(string(by))

	var metrics map[string]any
	err = json.NewDecoder(resp.Body).Decode(&metrics)
	if err != nil {
		fmt.Println("unable to decode response body", err)
		return
	}

	fmt.Println("debug metrics response", metrics)
}
