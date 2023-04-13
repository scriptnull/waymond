---
sidebar_position: 6
---

# Contribute

## Pre-requisites

1. [Go](https://go.dev/) v1.19 or above
1. [Just](https://github.com/casey/just)
1. [Node.js](https://nodejs.org/en) 16 or above (optional, only needed for building documentation site)

## Using nix

waymond repo contains [nix](https://nixos.org/) package manager configuration to enable contributiors to easily get all the needed software for development.

```sh
# clone waymond repo

$ cd waymond

$ nix-shell

# happy hacking!
```

## Build and run

```sh
$ just build

# waymond binary should be ready for use.

$ ./waymond
```

## Develop Docs

```sh
$ just run-site

# runs the website locally

$ just deploy

# deploys this site to GitHub pages
```