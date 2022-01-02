## Setup

Either git clone [the repo](https://github.com/itzg/mc-monitor) or [download a zip of the latest files](https://github.com/itzg/mc-monitor/archive/refs/heads/master.zip).

Go into the `examples/mc-monitor-prom` directory and start the composition using:

```shell
docker-compose up -d
```

Use `docker-compose logs -f` to watch the logs of the containers. When the Minecraft server with service name `mc` is up and running, then move onto the next step.

## Accessing Grafana

Open a browser to <http://localhost:3000> and log in initially with the username "admin" and password "admin". You will then be prompted to create a new password for the admin user.

[A dashboard called "MC Monitor"](http://localhost:3000/d/PpzSgJAnk/mc-monitor?orgId=1) is provisioned and can be accessed [in the dashboards section](http://localhost:3000/dashboards). That dashboard uses the Prometheus [datasource](http://localhost:3000/datasources) that was provisioned.

