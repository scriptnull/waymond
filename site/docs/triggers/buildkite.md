---
sidebar_position: 3
---

# buildkite

Trigger events for buildkite jobs waiting to be run on various buildkite queues.

```toml
[[trigger]]
type = "buildkite"
id = "my_buildkite_org"
# set BUILDKITE_TOKEN environment variable
```