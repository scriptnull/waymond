---
sidebar_position: 1
---

# Install

Download the binary for your OS and architecture from the [latest release](https://github.com/scriptnull/waymond/releases). Extract the compressed package and move it an executable PATH.

```sh
# example: for linux amd64
wget https://github.com/scriptnull/waymond/releases/latest/download/waymond_Linux_x86_64.tar.gz
tar -xf waymond_Linux_x86_64.tar.gz
mv waymond /usr/bin/waymond
```

## Run
The only way to run waymond right now is

```sh
waymond -config waymond.toml
```

But the project is looking to improve the CLI experience. So, please take a look [here](https://github.com/scriptnull/waymond/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fcli) if you would like to contribute.