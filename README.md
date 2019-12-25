
[![Docker Pulls](https://img.shields.io/docker/pulls/itzg/mc-monitor)](https://hub.docker.com/r/itzg/mc-monitor)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/itzg/mc-monitor)](https://github.com/itzg/mc-monitor/releases/latest)
[![CircleCI](https://circleci.com/gh/itzg/mc-monitor.svg?style=svg)](https://circleci.com/gh/itzg/mc-monitor)

Command/agent to monitor the status of Minecraft servers

## Install module

```
go get github.com/itzg/go-mc-status
```

## Usage

```
Subcommands:
	help             describe subcommands and their syntax
	status           Retrieves and displays the status of the given Minecraft server

Subcommands for monitoring:
	gather-for-telegraf  Periodically gathers to status of one or more Minecraft servers and sends metrics to telegraf over TCP using Influx line protocol
```

## Examples

### Checking the status of a server

```
docker run -ti --rm itzg/mc-monitor status --host mc.hypixel.net
```

where exit code will be 0 for success or 1 for failure.

### Monitoring a server

> The following example is provided in [examples/mc-monitor-telegraf](examples/mc-monitor-telegraf)

Given the telegraf config file:

```toml
[[inputs.socket_listener]]
  service_address = "tcp://:8094"

[[outputs.file]]
  files = ["stdout"]
```

...and a Docker composition of telegraf and mc-monitor services:

```yaml
version: '3'

services:
  telegraf:
    image: telegraf:1.13
    volumes:
    - ./telegraf.conf:/etc/telegraf/telegraf.conf:ro
  monitor:
    image: itzg/mc-monitor
    command: gather-for-telegraf
    environment:
      GATHER_INTERVAL: 10s
      GATHER_TELEGRAF_ADDRESS: telegraf:8094
      GATHER_SERVERS: mc.hypixel.net
```

The output of the telegraf service will show metric entries such as:

```
minecraft_status,host=mc.hypixel.net,port=25565,status=success response_time=0.172809649,online=51201i,max=90000i 1576971568953660767
minecraft_status,host=mc.hypixel.net,port=25565,status=success response_time=0.239236074,online=51198i,max=90000i 1576971579020125479
minecraft_status,host=mc.hypixel.net,port=25565,status=success response_time=0.225942383,online=51198i,max=90000i 1576971589006821324
```