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
# nats topics
### Local NATS subjects
- <app_id>.get.points '{ "scope": "all", "filter": { "tag": "abc" } }'
- <app_id>.get.points '{ "scope": "one",  "uuid": "<point_uuid>" }'
- <app_id>.get.points '{ "scope": "one", "name": "<point_name>" }'
- <app_id>.delete.points '{ "scope": "one", "uuid": "<point_uuid>" }'
- <app_id>.delete.points '{ "scope": "multiple", "uuids": ["<point_uuid1>", "<point_uuid2>"] }'
- <app_id>.post.points '{ "scope": "one", "name": "point 1", ... }'
- <app_id>.post.points '{ "scope": "multiple", "body": [{...}, {...}] }'
- <app_id>.put.points '{ "scope": "one", "name": "point 1", ... }'

### Cloud NATS to Edge Apps

- <global_uuid>.proxy.<app_id>.get.points '{ "scope": "all", "filter": { "tag": "abc" } }'
- <global_uuid>.proxy.<app_id>.get.points '{ "scope": "one",  "uuid": "<point_uuid>" }'
- <global_uuid>.proxy.<app_id>.get.points '{ "scope": "one", "name": "<point_name>" }'
- <global_uuid>.proxy.<app_id>.delete.points '{ "scope": "one", "uuid": "<point_uuid>" }'
- <global_uuid>.proxy.<app_id>.delete.points '{ "scope": "multiple", "uuids": ["<point_uuid1>", "<point_uuid2>"] }'
- <global_uuid>.proxy.<app_id>.post.points '{ "scope": "one", "name": "point 1", ... }'
- <global_uuid>.proxy.<app_id>.post.points '{ "scope": "multiple", "body": [{...}, {...}] }'
- <global_uuid>.proxy.<app_id>.put.points '{ "scope": "one", "name": "point 1", ... }'

### Cloud NATS to Edge BIOS
- <global_uuid>.get.system.ping
- <global_uuid>.get.systemctl.status '{ "service": "nubeio-rubix-os" }'
- <global_uuid>.post.apps.install '{ "name": "nubeio-rubix-os", "version":"v1.1" }'

# downloads

## nats server

```
https://nats.io/download/
```

The simplest way to just get the binary of a release of nats-server for your machine is to use the following shell command.

For example to get the binary for version 2.10.14 you would use:

```
curl -sf https://binaries.nats.dev/nats-io/nats-server/v2@v2.10.14 | sh
```

## nats cli
```
https://github.com/nats-io/natscli
```

### Installation from the shell
The following script will install the latest version of the nats cli on Linux and macOS:

```
curl -sf https://binaries.nats.dev/nats-io/natscli/nats@latest | sh
```

# nats servers
Start 2x nats servers 

server 1 (cloud-broker) (this is for the bios)
```
./nats-server --js --port 4222
```
server 2 (local-broker) (this is for ROS/FLEX and all modules)
```
./nats-server --js --port 4223
```

# start golang apps

## start ros/flex server (nats-local broker)
```
go run main.go --auth=false --uuid=abc --natsModulePort=4223
```

## start module (nats-local broker) (you need UWF installed)

`sudo` is need for `ufw` if you're not using `ufw` run as non root
```
cd modules/module-abc
go build main.go && sudo ./main
```

## start bios (nats-cloud broker)
```
cd modules/bios
go run *.go
```

# example commands
All commands are sent via the cloud broker

## BIOS commands (non proxy command)
```
./nats req "bios.<GLOBAL-UUID>.command" '{"command": "read_file", "body": {"path": "/home/aidan/test.txt"}}'
```

## Module command (is a proxy command)

open firewall port on module "my-module" if you have `ufw`
```
./nats req host.abc.modules.my-module "{\"command\": \"ufw\", \"body\": {\"subCommand\": \"open\", \"port\": 8080}}"
```
or get the time/date
```
./nats req host.abc.modules.my-module "{\"command\": \"ufw\", \"body\": {\"subCommand\": \"time\"}}"
```

## ROS/FLEX command (is a proxy command)

get all the hosts via RQL
```
./nats req host.abc.flex.rql "{\"script\": \"ctl.SystemdStatus(\\\"mosquitto\\\")\"}"
```

# using nats store

## add an object to the store

```
./nats object put mystore /home/aidan/test.txt 
```

## list all the stores
```
./nats req "bios.abc.store" '{"command": "get_stores"}''
```


## get all objects from a store
```
./nats req "bios.abc.store" '{"command": "get_store_objects", "body": {"store": "mystore"}}'
```