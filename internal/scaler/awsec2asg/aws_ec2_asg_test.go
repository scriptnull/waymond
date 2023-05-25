package awsec2asg

import (
	"context"
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
			config: `
id = "simple-asg-config"
minSize = 1
maxSize = 5
`,
			shouldErr: false,
		},
	}

	for idx, tc := range tt {
		k := koanf.New(".")
		err := k.Load(rawbytes.Provider([]byte(tc.config)), toml.Parser())
		if err != nil {
			t.Fatal("error while loading configuration")
		}
		trigger, err := ParseConfig(k)
		if err != nil {
			t.Fatal("invalid test case config. Pass a valid value for config")
		}

		func() {
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
