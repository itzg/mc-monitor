A library and command to check the status of a Minecraft server

Implements the handshake and status request from https://wiki.vg/Protocol

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
