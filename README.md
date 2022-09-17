# minecraft servers in AWS
This project manages different minecraft servers in AWS.
It stores them in S3 and runs them in ECS FARGATE.

Servers run as ECS FARGATE tasks.
* `business/containers/task.go` implements a template for creating these tasks. It takes a template struct as input,
which is loaded from a file on S3.
Then a task is created with the rendered template.
* `business/containers/status.go` implements getting the status of the current running task
* `business/containers/manage.go` implements the starting and stopping of the tasks
* `business/containers/rcon.go` implements sending commands to a running server

The task that is started has the following containers:
* setup - downloads the archive from S3 and extracts it in the shared volume
* main - runs the minecraft server from the shared volume; depends on setup completing successfully
* backup - runs backups on an interval; depends on the main container being started
* teardown - creates an archive from the shared volume and uploads it; depends on the main container being in a completed state.
