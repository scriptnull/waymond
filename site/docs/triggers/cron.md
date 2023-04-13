---
sidebar_position: 2
---

# Cron

Periodically trigger events in waymond.

```toml
[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"
```