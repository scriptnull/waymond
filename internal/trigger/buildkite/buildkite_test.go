package buildkite

import (
	"context"
	"os"
	"testing"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

func TestParseConfig(t *testing.T) {
	tt := []struct {
		config    string
		shouldErr bool
	}{
		{
			config:    "",
			shouldErr: true,
		},
		{
			config:    `id = "example-a"`,
			shouldErr: false,
		},
	}

	for _, tc := range tt {
		k := koanf.New(".")
		k.Load(rawbytes.Provider([]byte(tc.config)), toml.Parser())
		_, err := ParseConfig(k)
		switch {
		case tc.shouldErr && err == nil:
			t.Error("expected to error, but did not error")
		case !tc.shouldErr && err != nil:
			t.Errorf("expected not to error, but errored out with: %s", err)
		default:
			continue
		}
	}
}

func TestRegister(t *testing.T) {
	tt := []struct {
		buildkiteTokenEnvValue string
		config                 string
		shouldErr              bool
	}{
		{
			buildkiteTokenEnvValue: "",
			config:                 `id = "example-a"`,
			shouldErr:              true,
		},
		{
			buildkiteTokenEnvValue: "invalid-buildkite-token-value",
			config:                 `id = "example-a"`,
			shouldErr:              true,
		},
		{
			buildkiteTokenEnvValue: os.Getenv("TEST_BUILDKITE_TOKEN"),
			config:                 `id = "example-a"`,
			shouldErr:              false,
		},
	}

	for idx, tc := range tt {
		k := koanf.New(".")
		k.Load(rawbytes.Provider([]byte(tc.config)), toml.Parser())
		trigger, err := ParseConfig(k)
		if err != nil {
			t.Fatal("invalid test case config. Pass a valid value for config")
		}

		func() {
			os.Setenv(tokenEnvName, tc.buildkiteTokenEnvValue)
			defer os.Unsetenv(tokenEnvName)

			err = trigger.Register(context.Background())
			switch {
			case tc.shouldErr && err == nil:
				t.Errorf("[test case: %d] expected to error, but did not error", idx)
			case !tc.shouldErr && err != nil:
				t.Errorf("[test case: %d] expected not to error, but errored out with: %s", idx, err)
			}
		}()
	}
}
