#!/bin/bash

MEMINFO=`cat /proc/meminfo`
MEM_FREE=`echo $MEMINFO | awk '/MemFree/ { printf "%d", $2 }'`
MEM_TOTAL=`echo $MEMINFO | awk '/MemTotal/ { printf "%dn", $2 }'`

cat <<-INI
machine.hostname=$HOSTNAME
machine.mem_free=$MEM_FREE
machine.mem_total=$MEM_FREE
machine.load_avg=`cut -d" " -f1 /proc/loadavg`
machine.cpu_count=`grep -c 'cpu[0-9]' /proc/stat`
INI
