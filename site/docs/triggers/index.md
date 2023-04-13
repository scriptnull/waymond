---
hide_table_of_contents: true
---

# Triggers

Triggers are the components which could trigger an event, which could ultimately yield to autoscaling. The decision of when and how to auto-scale will flow from here.

| type | status | description |
|------|--------|-------------|
| cron | Available | Trigger events based on [cron expressions](https://en.wikipedia.org/wiki/Cron) |
| buildkite | Available | Trigger event based on the CI job queue length in Buildkite |

Propose a new trigger [here](https://github.com/scriptnull/waymond/issues/new).