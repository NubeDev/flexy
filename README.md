# flexy

Demo set up

```
  +-------------------+                    +--------------------+
  |  nats:4222        | <----------------->|      bios          |
  |  (cloud Broker)   |                    +--------------------+
  +-------------------+                             ^
                                                    |
                                                    v
                                           +--------------------+
                                           |  nats:4223         |
                                           |  (local Broker)    |
                                           +--------------------+
                                               ^         ^
                                               v         v
                                           +-------+   +----------------+
                                           |  ros  |   | ufw-module     |
                                           +-------+   | (moduleID:     |
                                                       |  my-module)    |
                                                       +----------------+

```

## nats servers
Start 2x nats servers 

server 1 (cloud-broker) (this is for the bios)
```
./nats-server --js --port 4222
```
server 2 (local-broker) (this is for ROS/FLEX and all modules)
```
./nats-server --js --port 4223
```

## start golang apps

### start ros/flex server (nats-local broker)
```
go run main.go --auth=false --uuid=abc --natsModulePort=4223
```

### start module (nats-local broker) (you need UWF installed)

`sudo` is need for `ufw` if you're not using `ufw` run as non root
```
cd modules/module-abc
go build main.go && sudo ./main
```

### start bios (nats-cloud broker)
```
cd modules/bios
go run *.go
```

## example commands
All commands are sent via the cloud broker

### BIOS commands (non proxy command)
```
./nats req "bios.<GLOBAL-UUID>.command" '{"command": "read_file", "body": {"path": "/home/aidan/test.txt"}}'
```

### Module command (is a proxy command)

open firewall port on module "my-module" if you have `ufw`
```
./nats req host.abc.modules.my-module "{\"command\": \"ufw\", \"body\": {\"subCommand\": \"open\", \"port\": 8080}}"
```
or get the time/date
```
./nats req host.abc.modules.my-module "{\"command\": \"ufw\", \"body\": {\"subCommand\": \"time\"}}"
```

### ROS/FLEX command (is a proxy command)

get all the hosts via RQL
```
./nats req host.abc.flex.rql "{\"script\": \"ctl.SystemdStatus(\\\"mosquitto\\\")\"}"
```