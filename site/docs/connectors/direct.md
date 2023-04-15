---
sidebar_position: 1
---

# Direct

Directly pass events between Trigger-Scaler or Trigger-Trigger.

## Configuration

For connecting trigger to scaler:

```toml
[[connect]]
type = "direct"
id = "<id>"
from = "trigger.<id>"
to = "scaler.<id>"
```

For connecting trigger to trigger:

```toml
[[connect]]
type = "direct"
id = "<id>"
from = "trigger.<id>"
to = "trigger.<id>"
```

Use `transform` block to transform the event data:

```toml
[connect.transform]
method = "<method>" # available method(s): "go_template"
template = """
{
  "some_other_field": "{{ .some_field }}",
}
"""
```

## Example

The following waymond config would periodically trigger the `buildkite` trigger. The `direct` connector with the `transform` block will make use of the output of `buildkite` trigger event and pass it to the specified Go template to produce the new data that will be sent to the destination of the connector.

```toml
[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"

[[trigger]]
type = "buildkite"
id = "my_buildkite_org"
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
"""
```

## Events

Emits `connector.%s.error` if there were any errors while sending the event data between the source and destination of the `direct` connector.