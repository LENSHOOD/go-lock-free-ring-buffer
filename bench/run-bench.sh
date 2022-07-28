#!/bin/bash

### init charts file
export LFRING_BENCH_CHARTS_FILE=table_define.dat
cat /dev/null > $LFRING_BENCH_CHARTS_FILE

### bench with capacity
export LFRING_BENCH_THREAD_NUM=12
export LFRING_BENCH_PRODUCER_NUM=6
echo \#title=with capacity\(threads=$LFRING_BENCH_THREAD_NUM, producers=$LFRING_BENCH_PRODUCER_NUM\),xAxis=capacity,yAxis=handover counts >> $LFRING_BENCH_CHARTS_FILE

capacities=(2 4 8 16 32 64 128 256 512 1024)
for cap in "${capacities[@]}"
do
   export LFRING_BENCH_CAP=$cap
   go test -run "^$" -bench "^.+MPMC$" -benchtime=10s -count=10 | \
    awk -v cap="$cap" '/NodeMPMC.+handovers/ || /HybridMPMC.+handovers/ || /ChannelMPMC.+handovers/ { printf "%s=(%d,%f)\n", $1, cap, $5 }' \
    >> $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

### bench with threads
export LFRING_BENCH_CAP=32
echo \#title=with thread number\(capacity=$LFRING_BENCH_CAP, producers=0.5*threads\),xAxis=threads,yAxis=handover counts >> $LFRING_BENCH_CHARTS_FILE

threads=(2 4 8 12 24 48)
for t in "${threads[@]}"
do
   export LFRING_BENCH_THREAD_NUM=$t
   export LFRING_BENCH_PRODUCER_NUM=$((t/2))
   go test -run "^$" -bench "^.+MPMC$" -benchtime=10s -count=10 | \
    awk -v thread="$t" '/NodeMPMC.+handovers/ || /HybridMPMC.+handovers/ || /ChannelMPMC.+handovers/ { printf "%s=(%d,%f)\n", $1, thread, $5 }' \
    >> $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

### bench with producer consumer ratio
export LFRING_BENCH_CAP=32
export LFRING_BENCH_THREAD_NUM=12
echo \#title=with producer\(capacity=$LFRING_BENCH_CAP, threads=$LFRING_BENCH_THREAD_NUM\),xAxis=producers,yAxis=handover counts >> $LFRING_BENCH_CHARTS_FILE

producers=(1 2 3 4 6 8 9 10 11)
for p in "${producers[@]}"
do
   export LFRING_BENCH_PRODUCER_NUM=$p
   go test -run "^$" -bench "^.+MPMC$" -benchtime=10s -count=10 | \
    awk -v producer="$p" '/NodeMPMC.+handovers/ || /HybridMPMC.+handovers/ || /ChannelMPMC.+handovers/ { printf "%s=(%d,%f)\n", $1, producer, $5 }' \
    >> $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

go test -run "^TestGenReport$"