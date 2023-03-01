package requester

import "testing"

func TestBuildkite(t *testing.T) {
	b := Buildkite{}
	b.Register()

	b.AutoScaleRequest()
}
