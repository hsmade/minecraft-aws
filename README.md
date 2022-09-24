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

## Secrets
* `ECR` name of the ECR (1234567890.dkr.ecr.<region>.amazonaws.com)
* `TFSTATE_BUCKET` name of the bucket
* `TFSTATE_REGION` name of the AWS region
* `TF_VAR_bucket` name of the bucket
* `TF_VAR_home_ip` IP used to restrict web UI to
* `TF_VAR_domain_name` domain name used for minecraft servers
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

## permissions needed for AWS user for CI:
* s3..

## Cost
### continuous costs
- route53 zone: 0.50 per month

## Todo
- define permissions above
- use iam role instead of user directly https://github.com/aws-actions/configure-aws-credentials#assuming-a-role
- current docker image will download the server and install it, every time. This takes too long
- ecs:
  - rclone config for backup
  - allow efs in ecs sg?
  - add user whitelist (env vars for server.properties?)
- web:
  - list servers (task definitions), with status (tasks)
  - stop task / stop server
    - run final backup
    - delete route53 record
  - create task / start server
    - create route53 record
  - rcon
