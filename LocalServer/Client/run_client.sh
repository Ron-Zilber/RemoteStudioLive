#!/bin/bash

if [ $# -eq 0 ]; then
    echo "Usage: $0 <ip_address>"
    exit 1
fi

ip_address="$1"
op_mode="record"
frame_size=480
setup="lab"
connType="tcp"

if [ $op_mode == "record" ]; then
    go run ClientUtils.go client.go $connType "$ip_address" 7777 $op_mode $frame_size "$@" 2>&1 | grep -v -E "ALSA lib|opus|silk|HarmShapeGain|~|Cannot connect to server socket|Cannot connect to server request channel|jack server is not running"
elif [ $op_mode == "song" ]; then
    go run ClientUtils.go client.go $connType "$ip_address" 7777 $op_mode 2>/dev/null | mpg123 -
fi  

python3 ./PlotGenerator.py ./Stats/StatisticsLog.txt ./Stats/interArrivalLog.txt $frame_size $setup $connType

#python3 ./multipleFrameSizePlotter.py $setup