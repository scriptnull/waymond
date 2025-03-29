package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/scriptnull/waymond/internal/connector"
	"github.com/scriptnull/waymond/internal/connector/direct"
)

const testcase1 = `[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"

[[trigger]]
type = "buildkite"
id = "my_buildkite_org"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable

[[connect]]
type = "direct"
id = "check_my_buildkite_org_queues_periodically"
from = "trigger.global_cron"
to = "trigger.my_buildkite_org"

[[scaler]]
type = "noop"
id = "noop"

[[connect]]
type = "direct"
id = "print_trigger_output"
from = "trigger.my_buildkite_org"
to = "scaler.noop"
[connect.transform]
method = "go_template"
template = """
{
  "asg_name": "{{ .queue }}",
  "desired_count": {{ .scheduled_jobs_count }}
}
"""`

const testcase2 = `
[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"

[[trigger]]
type = "buildkite"
id = "buildkite_1"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable

[[trigger]]
type = "buildkite"
id = "buildkite_2"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable

[[connect]]
type = "direct"
id = "connect_cron_to_buildkite_1"
from = "trigger.global_cron"
to = "trigger.buildkite_1"

[[connect]]
type = "direct"
id = "connect_buildkite_1_to_buildkite_2"
from = "trigger.buildkite_1"
to = "trigger.buildkite_2"

[[scaler]]
type = "noop"
id = "noop"

[[connect]]
type = "direct"
id = "connect_buildkite_1_to_noop"
from = "trigger.buildkite_1"
to = "scaler.noop"

[[connect]]
type = "direct"
id = "connect_buildkite_2_to_noop"
from = "trigger.buildkite_2"
to = "scaler.noop"`

const testcase3 = `
[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"

[[trigger]]
type = "buildkite"
id = "buildkite"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable

[[connect]]
type = "direct"
id = "connect_cron_to_buildkite"
from = "trigger.global_cron"
to = "trigger.buildkite"

[[connect]]
type = "direct"
id = "connect_buildkite_to_itself"
from = "trigger.buildkite"
to = "trigger.buildkite"`

const testcase4 = `
[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"

[[trigger]]
type = "buildkite"
id = "buildkite_1"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable

[[trigger]]
type = "buildkite"
id = "buildkite_2"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable

[[connect]]
type = "direct"
id = "connect_cron_to_buildkite_1"
from = "trigger.global_cron"
to = "trigger.buildkite_1"

[[connect]]
type = "direct"
id = "connect_buildkite_1_to_buildkite_2"
from = "trigger.buildkite_1"
to = "trigger.buildkite_2"

[[connect]]
type = "direct"
id = "connect_buildkite_2_to_buildkite_1"
from = "trigger.buildkite_2"
to = "trigger.buildkite_1"

[[scaler]]
type = "noop"
id = "noop"

[[connect]]
type = "direct"
id = "connect_buildkite_2_to_noop"
from = "trigger.buildkite_2"
to = "scaler.noop"`

func TestVerifyDAG(t *testing.T) {
	tests := []struct {
		name  string
		conf  string
		isDAG bool
	}{
		{
			name:  "cron-buildkite-noop",
			conf:  testcase1,
			isDAG: true,
		},
		{
			name:  "cron-buildkite1-noop, buildkite1-buildkite2, buildkite2-noop",
			conf:  testcase2,
			isDAG: true,
		},
		{
			name:  "cron-buildkite1-buildkite1",
			conf:  testcase3,
			isDAG: false,
		},
		{
			name:  "cron-buildkite1, buildkite1-buildkite2, buildkite2-noop, buildkite2-buildkite1",
			conf:  testcase4,
			isDAG: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			connectors, err := extractConnectors(tt.conf)
			if err != nil {
				t.Error("Exracting connectors returned errors")
			}
			got := verifyDAG(connectors)
			if got != tt.isDAG {
				t.Errorf("verifyDAG() = %v, want %v", got, tt.isDAG)
			}
		})
	}
}

func extractConnectors(conf string) (map[string]connector.Interface, error) {
	var kt = koanf.New(".")
	var errs []error

	// read waymond config file
	if err := kt.Load(rawbytes.Provider([]byte(conf)), toml.Parser()); err != nil {
		return nil, err
	}

	// track available connector configuration parsers available out of the box in waymond
	connectorConfigParsers := make(map[connector.Type]func(*koanf.Koanf) (connector.Interface, error))
	connectorConfigParsers[direct.Type] = direct.ParseConfig

	// extract connector from connector configurations
	connectorConfigs := kt.Slices("connect")
	connectors := make(map[string]connector.Interface)
	for _, connectorConfig := range connectorConfigs {
		ttype := connectorConfig.String("type")
		if ttype == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'type' field for connector: %+v", connectorConfig))
			continue
		}

		id := connectorConfig.String("id")
		if id == "" {
			errs = append(errs, fmt.Errorf("expected a non-empty 'id' field for connector: %+v", connectorConfig))
			continue
		}

		parseConfig, found := connectorConfigParsers[connector.Type(ttype)]
		if !found {
			errs = append(errs, fmt.Errorf("unknown 'type' value in connector: %s in %+v", ttype, connectorConfig))
			continue
		}

		connector, err := parseConfig(connectorConfig)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		connectors[id] = connector
	}
	if len(errs) > 0 {
		return nil, errors.New("Error Extracting connectors from the config file")
	}

	return connectors, nil
}
