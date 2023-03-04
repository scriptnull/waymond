<p align="center">
  <img src="https://user-images.githubusercontent.com/4211715/222185043-e82165e5-c755-4c4d-a10c-a28fad5503e7.png" height="128px">
  <br><br>
  <i>Autoscale Anything Anywhere All at once! :eyes:</i>
  <br>
</p>

&nbsp;

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

![image](https://user-images.githubusercontent.com/4211715/222920889-9d19dd0c-5d32-4ef2-918b-b5ac97b5845c.png)

## Contribute

**Software pre-requisites**

1. [Go](https://go.dev/) v1.19 or above
1. [Just](https://github.com/casey/just)

**Build**

Run `just build` and `./waymond` binary should be ready for use.
