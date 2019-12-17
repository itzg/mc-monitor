A library and command to check the status of a Minecraft server

Implements the handshake and status request from https://wiki.vg/Protocol

## Install module

```
go get github.com/itzg/go-mc-status
```

## Usage

With only `host` and `port` will run once and output the server info.

```
  -gather-interval duration
    	when gather endpoint configured, gathers and sends at this interval (env GATHER_INTERVAL) (default 1m0s)
  -gather-telegraf-address host:port
    	host:port of telegraf accepting Influx line protocol (env GATHER_TELEGRAF_ADDRESS)
  -host string
    	hostname of the Minecraft server (env MC_HOST) (default "localhost")
  -port int
    	port of the Minecraft server (env MC_PORT) (default 25565)
```
