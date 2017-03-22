# meteor [![Build Status](https://travis-ci.org/incu6us/meteor.svg)](https://travis-ci.org/incu6us/meteor) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/meteor-cd/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

The lightweight and quick continuous delivery tool 

![meteor](https://raw.githubusercontent.com/incu6us/meteor/master/examples/images/meteor.png)

### Purpose
It is very simple application in configuration and it is very quick. That's why it could be used on ARM's devices also, like RaspberryPI, to execute your `BASH` pipelines or scripts. Possibility, to integrate it with an external systems (for example: `TravisCI` and `Slack`) make it flexible. 

### Installation
##### build:
```
go build .
```

##### start command:
```
./meteor -conf meteor.conf
```

##### start with *systemd*:

```
mkdir /opt/meteor
cp -r {meteor,meteor.conf,tasks} /opt/meteor/
cp examples/systemd/meteor.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable meteor
systemctl start meteor
```

### General Configuration:
Main configuration for the service are placed in `meteor.conf`. Also, there is an additional configuration for each `task`, which is placed in `./tasks/` directory.  

General configuration file contains of:

`listen` - host and port to listening(example: ":8080")

`username` - username for basic authentication
`password` - password for basic authentication

`cmd-interpreter` - main interpreter for pipelines (default: `/bin/bash`)
`cmd-flag` - interpreter's flag (default: `-c`)

`slack-token` - Slack's verification token, for integrate with Slack API