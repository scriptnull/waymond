---
sidebar_position: 1
---

# Introduction

waymond is
- An open-source autoscaler.
- Aiming to provide autoscaling for a wide variety of infrastructure.
- Modular and extensible.
- Built with Go.

## Motivation

There are a lot of autoscalers out there like [AWS EC2 Autoscaling](https://docs.aws.amazon.com/autoscaling/ec2/userguide/what-is-amazon-ec2-auto-scaling.html), [kubernetes autoscalers](https://github.com/kubernetes/autoscaler), [nomad autoscaler](https://github.com/hashicorp/nomad-autoscaler), etc. Most of them are suited towards specific type of targets (example: nomad autoscaler could only be used for nomad). There is a good deal of overlap between those autoscalers in-terms of what and how an autoscale can happen.

The original idea for waymond came up while trying to autoscale CI/CD workloads in self-hosted infrastructure. Example: Autoscale the number of CI agents running as systemd processes inside a big EC2 VM and when we run out of limits there, try to bring up new EC2 VMs that run one CI agent per machine for a given CI job queue. At the sametime, autoscale the agents running in a kubernetes cluster when jobs are arriving in a different CI job queue.

waymond tries to support a variety of autoscaling targets from operating system processes to kubernetes clusters and everything in-between like traditional VMs. One of the main goals of the project is to make it very easy to autoscale mixed type of targets. Truly anything and anywhere!