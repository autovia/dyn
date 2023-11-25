# dyn

**IMPORTANT:** This is not production-ready software. This project is in active development.

## Introduction

ECS compatible container service.

Supports Authorization Header (AWS Signature Version 4)

Tested with:
* aws-cli/2.13.30 or greater
* aws-sdk-go-v2 v1.22.1
* aws-sdk-ruby3/3.185.2

## Development setup

Configure aws cli

cat $HOME/.aws/config

```shell
[default]
endpoint_url=http://localhost:3000
```

cat $HOME/.aws/config

```shell
[default]
aws_access_key_id = user
aws_secret_access_key = password
region = us-east-1
```

Run server

```shell
go run main.go
```

ECS commands

```shell
aws ecs register-task-definition --cli-input-json file://example/register-task-definition.json

aws ecs list-task-definitions

aws ecs run-task --cli-input-json file://example/run-task.json
```

## License

[Apache License 2.0](https://github.com/autovia/flightdeck/blob/master/LICENSE)

----
_Copyright [Autovia GmbH](https://autovia.io)_