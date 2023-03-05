<p align="center">
  <img src="https://user-images.githubusercontent.com/4211715/222973222-134cc720-3d44-484f-b39c-e2aaeb070bda.svg" height="128px">
  <br><br>
  <i>Autoscale Anything Anywhere All at once! :eyes:</i>
  <br>
</p>

&nbsp;

[![Go Report Card](https://goreportcard.com/badge/github.com/scriptnull/waymond)](https://goreportcard.com/report/github.com/scriptnull/waymond) [![lint](https://github.com/scriptnull/waymond/actions/workflows/lint.yaml/badge.svg?branch=main)](https://github.com/scriptnull/waymond/actions/workflows/lint.yaml) ![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/scriptnull/waymond) ![GitHub](https://img.shields.io/github/license/scriptnull/waymond)

waymond is
- An open-source autoscaler.
- Aiming to provide autoscaling for a wide variety of infrastructure.
- Modular and extensible.
- Built with Go.


## Motivation

There are a lot of autoscalers out there like [AWS EC2 Autoscaling](https://docs.aws.amazon.com/autoscaling/ec2/userguide/what-is-amazon-ec2-auto-scaling.html), [kubernetes autoscalers](https://github.com/kubernetes/autoscaler), [nomad autoscaler](https://github.com/hashicorp/nomad-autoscaler), etc. Most of them are suited towards specific type of targets (example: nomad autoscaler could only be used for nomad). There is a good deal of overlap between those autoscalers in-terms of what and how an autoscale can happen.

The original idea for waymond came up while trying to autoscale CI/CD workloads in self-hosted infrastructure. Example: Autoscale the number of CI agents running as systemd processes inside a big EC2 VM and when we run out of limits there, try to bring up new EC2 VMs that run one CI agent per machine for a given CI job queue. At the sametime, autoscale the agents running in a kubernetes cluster when jobs are arriving in a different CI job queue.

waymond tries to support a variety of autoscaling targets from operating system processes to kubernetes clusters and everything in-between like traditional VMs. One of the main goals of the project is to make it very easy to autoscale mixed type of targets. Truly anything and anywhere!

## Architecture

![architecture](https://user-images.githubusercontent.com/4211715/222922530-fda823c7-1a72-4156-99ac-3d249e4e8e47.png)

## Concepts

### Triggers

Triggers are the components which could trigger an event, which could ultimately yield to autoscaling. The decision of when and how to auto-scale will flow from here.

### Scalers

Scalers are the components which perform autoscaling operation. They access the target APIs (like docker or AWS API) to achieve a desired state in those systems.

### Connectors

Connectors are the components which connect two objects to facilitate the flow of events between them. Example: A connector could connect a cron trigger to docker scaler. This will ensure that the desired number of docker containers are running in a machine periodically.

Connectors can connect a trigger to a trigger (and form a chain of triggers) that ultimately is connected to a scaler. Connectors are also a good place to do data transformation of data from a trigger to a format of data that a scaler can understand.

### Event Bus
All the components `triggers`, `scalers`, and `connectors` are internally connected via a simple event bus (don't be scared it is just a Go channel and some helper functions :smile:). This event-based architecture will help any of the above mentioned components to capture and act on events in a seamless way.

## Install

Download the binary for your OS and architecture from the [latest release](https://github.com/scriptnull/waymond/releases). Extract the compressed package and move it an executable PATH.

```sh
# example: for linux amd64
wget https://github.com/scriptnull/waymond/releases/latest/download/waymond_Linux_x86_64.tar.gz
tar -xf waymond_Linux_x86_64.tar.gz
mv waymond /usr/bin/waymond
```

## Command
The only way to run waymond right now is

```sh
waymond -config waymond.toml
```

But the project is looking to improve the CLI experience. So, please take a look [here](https://github.com/scriptnull/waymond/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fcli) if you would like to contribute.


## Configuration

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

## Modules

### Trigger

| type | status | description |
|------|--------|-------------|
| cron | Available | Trigger events based on [cron expressions](https://en.wikipedia.org/wiki/Cron) |
| http_endpoint | Looking for contribution | Starts a HTTP server in waymond and triggers event for every HTTP request |
| http_client | Looking for contribution | Creates a HTTP client in waymond and triggers event based on the HTTP response |
| buildkite | [Planned](https://github.com/scriptnull/waymond/milestone/1) | Trigger event based on the CI job queue length in Buildkite |

Propose a new trigger [here](https://github.com/scriptnull/waymond/issues/new).

### Scaler

| type | status | description |
|------|--------|-------------|
| docker | Available | Autoscales docker containers |
| docker_compose | Looking for contribution | Autoscales containers in a docker compose setup |
| aws_ec2 | Looking for contribution | Autoscales AWS EC2 machines |
| aws_ec2_fleet | Looking for contribution | Autoscales AWS EC2 machines via [EC2 Fleet](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-fleet.html) |
| aws_ec2_asg | [Planned](https://github.com/scriptnull/waymond/milestone/1) | Autoscales AWS EC2 machines via Autoscaling groups |

Propose a new scaler [here](https://github.com/scriptnull/waymond/issues/new).

### Connector

| type | status | description |
|------|--------|-------------|
| direct | Available | Directly pass events between Trigger-Scaler or Trigger-Trigger |
| regex | [Planned](https://github.com/scriptnull/waymond/milestone/1) | Like `direct`, but uses regular expressions to transform data sent between the components |

Propose a new connector [here](https://github.com/scriptnull/waymond/issues/new).

## Contribute

**Software pre-requisites**

1. [Go](https://go.dev/) v1.19 or above
1. [Just](https://github.com/casey/just)

**Build**

Run `just build` and `./waymond` binary should be ready for use.
