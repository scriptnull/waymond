---
hide_table_of_contents: true
---

# Triggers

Triggers are the components which could trigger an event, which could ultimately yield to autoscaling. The decision of when and how to auto-scale will flow from here.

| type | status | description |
|------|--------|-------------|
| cron | Available | Trigger events based on [cron expressions](https://en.wikipedia.org/wiki/Cron) |
| buildkite | [In progress](https://github.com/scriptnull/waymond/milestone/1) | Trigger event based on the CI job queue length in Buildkite |
| http_endpoint | Looking for contribution | Starts a HTTP server in waymond and triggers event for every HTTP request |
| http_client | Looking for contribution | Creates a HTTP client in waymond and triggers event based on the HTTP response |

Propose a new trigger [here](https://github.com/scriptnull/waymond/issues/new).