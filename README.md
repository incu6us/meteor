# meteor [![Build Status](https://travis-ci.org/incu6us/meteor.svg)](https://travis-ci.org/incu6us/meteor) [![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/meteor-cd/Lobby?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

The lightweight and quick continuous delivery tool 

![meteor](https://raw.githubusercontent.com/incu6us/meteor/master/examples/images/meteor.png)

[Image from clipartfest.com](https://img.clipartfest.com/faa6398d9fac10f8ad4a565532e62af4_fireball-20clipart-clipart-comet-clipart-gif_1600-1200.svg)

### Purpose
It is very simple application with configuration and it is very quick. That's why it could be used on ARM's devices also, like RaspberryPI, to execute your `BASH` pipelines or scripts. Possibility to integrate it with an external systems (for example: `TravisCI` and `Slack`) make it flexible. 

## Installation
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

#### General Configuration:
Main configuration for the service are placed in `meteor.conf`. Also, there is an additional configuration for each `task`, which is placed in `./tasks/` directory.  

General configuration file contains of:

`listen` - host and port to listening(example: ":8080")

`username` - username for basic authentication
`password` - password for basic authentication

`cmd-interpreter` - main interpreter for pipelines (default: `/bin/bash`)
`cmd-flag` - interpreter's flag (default: `-c`)

`slack-token` - Slack's verification token, for integration with Slack API

## Tasks
#### Task creation
To create a new task you just need to create a directory inside `./tasks/` and two files in the new created folder:
 
 ```
 mkdir ./task/some-new-task
 touch ./task/some-new-task/{config,pipeline}
 ```

`config` - additional configuration file for each task, which is basically used for Slack's webhooks integration, to get a messages from the task. 

Example:
```
[slack]
webhook-url = "https://hooks.slack.com/services/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"
```

`pipeline` - a pipeline chain

Example:
```
export VAR1="1"; echo $VAR1
export VAR1="1"; exaport VAR2="2"; echo $VAR1 $VAR2
```

*You don't need to reload the application after creation a task*

# Integration
There are a couple of http calls, which will help you to integrate it with an external systems.

#### API Calls:
- To execute a task:

    `/api/task/run/{taskName}?param1=var1` - API, for execution of a task, where `{taskName}` is a folder in tasks dir.
    You can use `username` and `password` from general configuration to turning on a basic authorization. 
    CURL example to execute a task with basic auth header:

    ```
    curl -i -H 'Authorization: Basic dXNlcjo2NjY2NjY=' 'http://localhost:8080/api/task/run/test?param1=var1'
    ```
    `param1` - is a parameter for pipeline via HTTP interface and could be used in you script(like: `$param1`). It is working only via HTTP-interface
    
    If you will configure a `webhook-url` for Slack, then you will be able to get a status messages from the call.

- To integrate with Slack:

    - `/api/integration/slack/list` - to list available tasks;
    - `/api/integration/slack/run`  - run task manually from Slack
    
  to integrate this calls you need to go to the `https://api.slack.com` and create a new application. Then, you need to get a verification token and put it into main configuration.
  ![slack token](https://raw.githubusercontent.com/incu6us/meteor/master/examples/images/slack_token.png)

  the second step, will be to create a "Slash commands": one for `list` and another for `run`
  ![slash commands](https://raw.githubusercontent.com/incu6us/meteor/master/examples/images/slash_commands.png)
  
  Examples:
  - taskrun
  ![taskrun](https://raw.githubusercontent.com/incu6us/meteor/master/examples/images/taskrun.png)
  
  - tasklist:
  ![taskrun](https://raw.githubusercontent.com/incu6us/meteor/master/examples/images/tasklist.png)