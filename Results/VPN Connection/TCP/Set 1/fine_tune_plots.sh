#!/bin/bash

frame_size=120
setup="home"
connType="tcp"

python3 ./PlotGenerator.py ./Stats/StatisticsLog.txt ./Stats/interArrivalLog.txt $frame_size $setup $connType

frame_size=240
 
python3 ./PlotGenerator.py ./Stats/StatisticsLog.txt ./Stats/interArrivalLog.txt $frame_size $setup $connType

frame_size=480

python3 ./PlotGenerator.py ./Stats/StatisticsLog.txt ./Stats/interArrivalLog.txt $frame_size $setup $connType

frame_size=960
   
python3 ./PlotGenerator.py ./Stats/StatisticsLog.txt ./Stats/interArrivalLog.txt $frame_size $setup $connType

frame_size=1920

python3 ./PlotGenerator.py ./Stats/StatisticsLog.txt ./Stats/interArrivalLog.txt $frame_size $setup $connType

python3 ./multipleFrameSizePlotter.py $setup $connType