---
sidebar_position: 3
---

# Configuration

waymond accepts a toml configuration file when it boots up to configure all its components.

```toml
#
# For every 1 minute, waymond will check if there are two redis docker containers running and run them if not
#

[[trigger]]
type = "cron"
id = "global_cron"
expression = "*/1 * * * *"

[[scaler]]
type = "docker"
id = "local_redis_containers"
image_name = "redis"
image_tag = "latest"
count = 2

[[connect]]
type = "direct"
id = "run_redis_via_cron"
from = "trigger.global_cron"
to = "scaler.local_redis_containers"
```