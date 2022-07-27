#!/bin/bash

### init charts file
export LFRING_BENCH_CHARTS_FILE=table_define.dat
cat /dev/null > $LFRING_BENCH_CHARTS_FILE

### bench with capacity
echo \#title=with capacity,xAxis=capacity,yAxis=ns/op >> $LFRING_BENCH_CHARTS_FILE

capacities=(2 4 8 16 32 64 128 256 512 1024)
for cap in "${capacities[@]}"
do
   export LFRING_BENCH_CAP=$cap
   go test -run "^$" -bench "^.+MPMC$" -benchtime=10s -count=5 | \
    benchstat /dev/stdin | \
    awk -v cap="$cap" '/NodeMPMC/ || /HybridMPMC/ || /ChannelMPMC/ { printf "%s=(%d,%f)\n", $1, cap, $2 }' \
    >> $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

go test -run "^TestGenReport$"