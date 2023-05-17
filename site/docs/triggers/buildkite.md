---
sidebar_position: 3
---

# Buildkite

Trigger events for buildkite jobs waiting to be run on various buildkite queues.

```toml
[[trigger]]
type = "buildkite"
id = "my_buildkite_org"
filter_by_queue_name = "aws-on-demand-arm64-ubuntu-.*-ami-.*"
# set BUILDKITE_TOKEN environment variable
```