<p align="center">
  <img src="https://user-images.githubusercontent.com/4211715/222973222-134cc720-3d44-484f-b39c-e2aaeb070bda.svg" height="128px" style="max-width: 100%;">
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
Most of the auto-scalers nowadays are very specific to particular use-cases. They fall into one or more of the following categories:

- Cloud specific
  - [AWS EC2 Autoscaling groups](https://docs.aws.amazon.com/autoscaling/ec2/userguide/what-is-amazon-ec2-auto-scaling.html)
  - [GCP Managed Instance groups](https://cloud.google.com/compute/docs/instance-groups)
- Platform-specific
  - [Kubernetes autoscaler](https://github.com/kubernetes/autoscaler)
  - [Nomad autoscaler](https://github.com/hashicorp/nomad-autoscaler)
  - [Karpeneter](https://karpenter.sh/)
  - [Keda](https://keda.sh/)
- Type and workload-specific
  - Horizontal scaling
  - Vertical scaling
  - Compatible with containers only
  - Compatible with VMs only
 
There is a good deal of overlap between those autoscalers in-terms of what and how an autoscale can happen. Waymond tries to support a variety of autoscaling targets from operating system processes to kubernetes clusters and everything in-between like traditional VMs. One of the main goals of the project is to make it very easy to autoscale mixed type of targets. Truly anything and anywhere!

The original idea for waymond came up while trying to autoscale CI/CD workloads in self-hosted infrastructure. Example: Autoscale the number of CI agents running as systemd processes inside a big EC2 VM and when we run out of limits there, try to bring up new EC2 VMs that run one CI agent per machine for a given CI job queue. At the sametime, autoscale the agents running in a kubernetes cluster when jobs are arriving in a different CI job queue.

## Architecture

![architecture](https://user-images.githubusercontent.com/4211715/222922530-fda823c7-1a72-4156-99ac-3d249e4e8e47.png)

## Concepts

### Triggers

Triggers are the components which could trigger an event, which could ultimately yield to autoscaling. The decision of when and how to auto-scale will flow from here.

| type          | status                   | description                                                                    |
| ------------- | ------------------------ | ------------------------------------------------------------------------------ |
| cron          | Available                | Trigger events based on [cron expressions](https://en.wikipedia.org/wiki/Cron) |
| http_endpoint | Looking for contribution | Starts a HTTP server in waymond and triggers event for every HTTP request      |
| http_client   | Looking for contribution | Creates a HTTP client in waymond and triggers event based on the HTTP response |
| buildkite     | Available                | Trigger event based on the CI job queue length in Buildkite                    |

Propose a new trigger [here](https://github.com/scriptnull/waymond/issues/new).

### Scalers

Scalers are the components which perform autoscaling operation. They access the target APIs (like docker or AWS API) to achieve a desired state in those systems.

| type           | status                   | description                                                                                                         |
| -------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------- |
| docker         | Available                | Autoscales docker containers                                                                                        |
| docker_compose | Looking for contribution | Autoscales containers in a docker compose setup                                                                     |
| noop           | Available                | `noop` stands for "No-Operation". This is mainly for debugging what data is being received by a problematic scaler. |
| aws_ec2        | Looking for contribution | Autoscales AWS EC2 machines                                                                                         |
| aws_ec2_fleet  | Looking for contribution | Autoscales AWS EC2 machines via [EC2 Fleet](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-fleet.html)     |
| aws_ec2_asg    | Available                | Autoscales AWS EC2 machines via Autoscaling groups                                                                  |

Propose a new scaler [here](https://github.com/scriptnull/waymond/issues/new).

### Connectors

Connectors are the components which connect two objects to facilitate the flow of events between them. Example: A connector could connect a cron trigger to docker scaler. This will ensure that the desired number of docker containers are running in a machine periodically.

Connectors can connect a trigger to a trigger (and form a chain of triggers) that ultimately is connected to a scaler. Connectors are also a good place to do data transformation of data from a trigger to a format of data that a scaler can understand.

| type   | status    | description                                                    |
| ------ | --------- | -------------------------------------------------------------- |
| direct | Available | Directly pass events between Trigger-Scaler or Trigger-Trigger |

Propose a new connector [here](https://github.com/scriptnull/waymond/issues/new).

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

## Contribute

**Software pre-requisites**

1. [Go](https://go.dev/) v1.19 or above
1. [Just](https://github.com/casey/just)
1. [Node.js](https://nodejs.org/en) 16 or above (optional, only needed for building documentation site)

If you are a user of [nix](https://nixos.org/) package manager, you can enter the `nix-shell` to automatically download all the above dependencies.

**Build**

Run `just build` and `./waymond` binary should be ready for use.

## Progress
waymond is still an ⚠️ alpha software ⚠️ at this point.

- July 2022: waymond deployment continues to run. Two new contributors for the project so far. [v0.2.2](https://github.com/scriptnull/waymond/releases/tag/v0.2.2) was released for fixing bugs. 🐛
- June 2022: waymond was deployed in a real-world use-case to autoscale hundreds of machines. [v0.2.1](https://github.com/scriptnull/waymond/releases/tag/v0.2.1) was released. 💯
- May 2022: waymond [v0.2.0](https://github.com/scriptnull/waymond/releases/tag/v0.2.0) with minimal features needed to run it in real-world workloads 👷‍♂️
- March 2022: waymond won a prize in [FOSS Hack 3.0](https://forum.fossunited.org/t/foss-hack-3-0-results/1882) 🏆🏅

## Community

#### Talk to a human

The project is currently in very early stages and it would be awesome if you could join us! If you are looking to talk to a human about the waymond project, feel free join the [telegram group](https://t.me/+SUoglr-nx2JhMmEy). If you are interested in requesting new features or report bugs, please do so in [the issue tracker](https://github.com/scriptnull/waymond/issues). If you are looking to contribute for the first time, try checking the [issues tagged with good-first-issue](https://github.com/scriptnull/waymond/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22).

#### Swag

All contributors to the waymond project can optionally choose to receive an one-time swag. If you have contributed to the project, please fill out [this form](https://forms.gle/cigHWuw6ypZSLnxPA) to opt-in for receiving the swag.
