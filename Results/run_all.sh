#!/bin/bash

cd VPN\ Connection/
echo "VPN Connection"
cd TCP
cd 'Set 1'
./fine_tune_plots.sh &
cd ..
cd 'Set 2'
./fine_tune_plots.sh &
cd ..
cd 'Set 3'
./fine_tune_plots.sh &
cd ..
cd ..

cd UDP
cd 'Set 1'
./fine_tune_plots.sh 
cd ..
cd 'Set 2'
./fine_tune_plots.sh &
cd ..
cd 'Set 3'
./fine_tune_plots.sh &
cd ..
cd ..
cd ..

echo "Wifi Connection"
cd Wifi\ connection\ -\ LAN/

cd TCP
cd 'Set 1'
./fine_tune_plots.sh 
cd ..
cd 'Set 2'
./fine_tune_plots.sh &
cd ..
cd 'Set 3'
./fine_tune_plots.sh &
cd ..
cd ..

cd UDP
cd 'Set 1'
./fine_tune_plots.sh 
cd ..
cd 'Set 2'
./fine_tune_plots.sh &
cd ..
cd 'Set 3'
./fine_tune_plots.sh &
cd ..
cd ..
cd ..

echo "Wired Connection"
cd Wired\ Connection/
cd TCP
cd 'Set 1'
./fine_tune_plots.sh 
cd ..
cd 'Set 2'
./fine_tune_plots.sh &
cd ..
cd 'Set 3'
./fine_tune_plots.sh &
cd ..
cd ..

cd UDP
cd 'Set 1'
./fine_tune_plots.sh &
cd ..
cd 'Set 2'
./fine_tune_plots.sh 
cd ..
cd 'Set 3'
./fine_tune_plots.sh 

