---
sidebar_position: 2
---

# cron

Periodically trigger events in waymond.

```toml
[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"
```