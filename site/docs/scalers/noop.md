---
sidebar_position: 2
---

# No-op

No-op stands for ["No operation"](https://en.wikipedia.org/wiki/NOP_(code)). No-op scaler is used for debugging purposes. It doesn't do anything other than logging the event data received from the triggers connected to it.


## Configuration

```toml
[[scaler]]
type = "noop"
id = "<choose a name>"
```

## Example

The following waymond config would periodically trigger the `buildkite` trigger. The events from buildkite trigger will not be used to perform any autoscaling operation. Instead, the noop scaler connected to it will log all the event data in waymond logs.

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
```

## Events

Emits `scaler.<id>.output` event that produces the same data received in the noop scaler as the input via `scaler.<id>.input` event.