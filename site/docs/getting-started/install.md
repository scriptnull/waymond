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

## Run as a Systemd Service
After downloading the binary to `/usr/bin` directory, create a new service unit file in the `/etc/systemd/system` directory named `waymond.service`:

```
[Unit]
Description=An Awesome Autoscaler Service
After=network.target

[Service]
ExecStart=/usr/bin/waymond -config /path/to/config/file
WorkingDirectory=/usr/bin
Restart=always
RestartSec=10
User=your_user
Group=your_group

[Install]
WantedBy=multi-user.target
```
Reload the systemd manager configuration to apply the new service: 
```sudo systemctl daemon-reload```

Enable the service to start on boot:

```sudo systemctl enable waymond.service```

Start the service:

```sudo systemctl start waymond.service```

But the project is looking to improve the CLI experience. So, please take a look [here](https://github.com/scriptnull/waymond/issues?q=is%3Aissue+is%3Aopen+label%3Aarea%2Fcli) if you would like to contribute.
