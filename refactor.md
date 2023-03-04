# Design
## EC2
* needs SSM `AmazonSSMManagedInstanceCore`
* needs access to EFS `AmazonElasticFileSystemClientReadWriteAccess`
* uses https://github.com/itzg/mc-monitor/releases/download/0.11.0/mc-monitor_0.11.0_linux_amd64.tar.gz
* run bootscript
  * install java `apt install default-jre-headless --no-install-recommends -y`
  * start minecraft from script in tags on EFS
  * wait for minecraft to start; terminate server when no users for x mins
    * `mc-monitor status --host localhost --port "$SERVER_PORT" --show-player-count`

## EFS filesystem
* tagged with DNS name and start script path
* contains properties, server install, plugins and world
* aws_efs_mount_target
* aws_efs_file_system

## Flows
### Create server
* create EFS
* create SG for EFS; attach to EFS
* Start EC2 VM
  * SG minecraft
  * no minecraft loaded, so no auto-stop

### List servers
* get EFS's

### Start server
* start EC2 VM with EFS mounted
  * attach EFS
  * minecraft SG
  * terminate on stop
  * ssm
  * subnet
  * 4GB
* tag with name from EFS tags
* create DNS record

### Stop server
* terminal EC2 VM
* remove DNS record

### Status
* list running VMs with name
  * pending
  * started
  * mc starting
  * mc running
  * shutting down
  * terminated


# mist
## security groups
### ingress
* 25565 minecraft port
* 8080 metrics (filtered)

### egress
* all to all

## bootscript
```bash
#cloud-boothook
#!/bin/bash

cat << EOF > /start.sh
#!/bin/bash
max_startup=900
initial_players_timeout=900
idle_time_timeout=900

apt update
apt-get -y install --no-install-recommends nfs-common wget tar gzip openjdk-19-jre-headless screen
wget https://github.com/itzg/mc-monitor/releases/download/0.11.0/mc-monitor_0.11.0_linux_amd64.tar.gz -O - | tar xzvf - -C /usr/local/bin/
file_system_id_1=fs-00d665d419d8c44ad
mount -t nfs4 -o vers=4.1 \$(curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone).\${file_system_id_1}.efs.\$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document|grep region|awk -F\" '{print \$4}').amazonaws.com:/ /mnt
cd /mnt && screen -d -m ./run.sh 
mc-monitor  export-for-prometheus -servers localhost &

boot=\$(date +%s)
while true
do 
  mc-monitor status && break
  now=\$(date +%s)
  (( waiting = now - boot ))
  if [ \${waiting} -gt \${max_startup} ]
  then
    echo "Timeout waiting for minecraft to start, stopping server"
    halt
  fi
  sleep 1
done

start=\$(date +%s)
last_online=0
while sleep 1
do
  now=\$(date +%s)
  player_count=\$(mc-monitor status -show-player-count)
  if [ \${player_count} -gt 0 ]
  then
    echo "\${player_count} players online"
    last_online=\${now}
    continue 
  fi 
  
  if [ \${last_online} -eq 0 ]
  then
    echo "waiting for first player(s) to come online"
    (( initial_players_time = now - start ))
    if [ \${initial_players_time} -gt \${initial_players_timeout} ]
    then
      echo "waiting for initial players timed out at \${initial_players_time} seconds"
      break
    fi
    continue 
  fi
  
  if [ \${player_count} -eq 0 ]
  then
    (( idle_time = now - last_online ))
    echo "no players online for \${idle_time} seconds"
    if [ \${idle_time} -gt \${idle_time_timeout} ]
    then
      echo "idle timeout, stopping server"
      break
    fi
  fi
done

halt
EOF
chmod +x /start.sh
screen -d -m /start.sh

```

## Todo
* create new server (EFS) with efs SG attached https://docs.aws.amazon.com/efs/latest/ug/wt1-create-efs-resources.html
