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
