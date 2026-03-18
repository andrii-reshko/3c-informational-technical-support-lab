#!/usr/bin/env bash

min_start=100000000
max_start=5000000000
min_mid=$((min_start * 10))
max_mid=$((max_start * 10))
min_end=$((min_start * 100))
max_end=$((max_start * 100))

node_map=(
  "node-1:9101"
  "node-2:9102"
)

ramp_up=30
steady=60
ramp_down=30
total=$((ramp_up + steady + ramp_down))

echo "Load test: ${ramp_up}s ramp-up, ${steady}s steady, ${ramp_down}s ramp-down"

for ((i = 1; i <= total; i++)); do
  if [ "$i" -le "$ramp_up" ]; then
    progress=$((i * 100 / ramp_up))
    min_cpu_ops=$min_start
    max_cpu_ops=$max_start
    echo ">>> [${i}/${total}] RAMP-UP ${progress}% - warming up"
  elif [ "$i" -le $((ramp_up + steady)) ]; then
    min_cpu_ops=$min_mid
    max_cpu_ops=$max_mid
    echo ">>> [${i}/${total}] STEADY - max load"
  else
    remaining=$((total - i))
    progress=$((remaining * 100 / ramp_down))
    min_cpu_ops=$min_start
    max_cpu_ops=$max_start
    echo ">>> [${i}/${total}] RAMP-DOWN ${progress}% - cooling"
  fi

  for node in "${node_map[@]}"; do
    IFS=':' read -r node_name node_port <<< "$node"
    range=$((max_cpu_ops - min_cpu_ops + 1))
    seed=$(od -An -tu8 -N8 /dev/urandom | tr -d ' ')
    cpu=$(((seed % range) + min_cpu_ops))
    curl -s -o /dev/null "http://localhost:${node_port}?cpu=${cpu}&delay=0" &
  done
  sleep 1
done

echo "DONE"