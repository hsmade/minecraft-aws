#cloud-boothook
#!/bin/bash

cat << EOF > /start.sh
#!/bin/bash
max_startup=900
initial_players_timeout=900
idle_time_timeout=900

apt update
apt-get -y install --no-install-recommends nfs-common wget tar gzip openjdk-17-jre-headless screen
wget https://github.com/itzg/mc-monitor/releases/download/0.11.0/mc-monitor_0.11.0_linux_amd64.tar.gz -O - | tar xzvf - -C /usr/local/bin/
mc-monitor  export-for-prometheus -servers localhost &

file_system_id=FSID # gets replaced with the actual ID
mount -t nfs4 -o vers=4.1,exec \${file_system_id}.efs.\$(curl -s http://169.254.169.254/latest/dynamic/instance-identity/document|grep region|awk -F\" '{print \$4}').amazonaws.com:/ /mnt
cd /mnt && screen -d -m ./run.sh

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
    (( initial_players_time = now - start ))
    echo "[${initial_players_time}s] waiting for first player(s) to come online"
    if [ \${initial_players_time} -gt \${initial_players_timeout} ]
    then
      echo "[${initial_players_time}s] waiting for initial players timed out"
      break
    fi
    continue
  fi

  if [ \${player_count} -eq 0 ]
  then
    (( idle_time = now - last_online ))
    echo "[${idle_time}s] no players online..."
    if [ \${idle_time} -gt \${idle_time_timeout} ]
    then
      echo "[${idle_time}s] idle timeout, stopping server"
      break
    fi
  fi
done

halt
EOF
chmod +x /start.sh
screen -d -m /start.sh
