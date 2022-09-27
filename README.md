# minecraft servers in AWS
This project manages different minecraft servers in AWS.
It stores them in EFS and runs them in ECS FARGATE.

The server definitions are implemented as ECS task definitions.
Running them creates a task in ECS, and assigns a dns record to the public IP.
When you shut a server down, the DNS record is removed, and the task stopped.

There is a backup process that backs up to S3, with retention
This backup only backs things up if there are users in the server.
When all users leave, the server will shut down after the defined period.

## Secrets
* `ECR` name of the ECR (1234567890.dkr.ecr.<region>.amazonaws.com)
* `TFSTATE_BUCKET` name of the bucket where the terraform state is stored (created manually)
* `TFSTATE_REGION` name of the AWS region
* `TF_VAR_home_ip` IP used to restrict web UI to
* `TF_VAR_domain_name` domain name used for minecraft servers
* `TF_VAR_whitelist` minecraft users to whitelist
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`

## permissions needed for AWS user for CI:
* s3..

## Cost
### continuous costs
- route53 zone: 0.50 per month
- efs

## Todo
- define permissions above
- use iam role instead of user directly https://github.com/aws-actions/configure-aws-credentials#assuming-a-role
- current docker image will download the server and install it, every time. This takes too long
- infra:
  - rclone config for backup
  - come up with a way to load servers from file/secret, in TF
  - add SG for ECS for EFS (now uses some default?)
  - limit iam:PassRole
  - delete task: `deleting DNS record: getting IP: timeout waiting for server to get IP`
  - add optional firewall for public IP
  - fix soa record
  - have a sidecar upload the thumbnail for a server?
- web:
  - list servers (task definitions), with status (tasks) and tags
  - stop task / stop server
    - run final backup
    - delete route53 record
  - create task / start server
    - create route53 record
  - rcon
- pipeline:
  - lamdba depends on image, image depends on ecr -> deadlock