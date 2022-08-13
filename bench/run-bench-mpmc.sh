#!/bin/bash

do_bench() {
  go test -run "^$" -bench "^.+MPMC$" -benchtime=1s -count=100 | \
    awk -v value="$1" '/NodeMPMC.+handovers/ || /HybridMPMC.+handovers/ || /ChannelMPMC.+handovers/ { printf "%s=(%d,%f)\n", $1, value, $5 }' \
    >> "$2"
}

### init charts file
export LFRING_BENCH_CHARTS_FILE=table_define.dat
cat /dev/null > $LFRING_BENCH_CHARTS_FILE

### bench with capacity
echo \#title=with capacity\(threads=$LFRING_BENCH_THREAD_NUM, producers=$LFRING_BENCH_PRODUCER_NUM\),xAxis=capacity,yAxis=handover counts >> $LFRING_BENCH_CHARTS_FILE

export LFRING_BENCH_THREAD_NUM=12
export LFRING_BENCH_PRODUCER_NUM=6
capacities=(2 4 8 16 32 64 128 256 512 1024)
for cap in "${capacities[@]}"
do
   export LFRING_BENCH_CAP=$cap
   do_bench "$cap" $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

### bench with threads
echo \#title=with thread number\(capacity=$LFRING_BENCH_CAP, producers=0.5*threads\),xAxis=threads,yAxis=handover counts >> $LFRING_BENCH_CHARTS_FILE

export LFRING_BENCH_CAP=32
threads=(2 4 8 12 24 48)
for t in "${threads[@]}"
do
   export LFRING_BENCH_THREAD_NUM=$t
   export LFRING_BENCH_PRODUCER_NUM=$((t/2))
   do_bench "$t" $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

### bench with producer consumer ratio
echo \#title=with producer\(capacity=$LFRING_BENCH_CAP, threads=$LFRING_BENCH_THREAD_NUM\),xAxis=producers,yAxis=handover counts >> $LFRING_BENCH_CHARTS_FILE

export LFRING_BENCH_CAP=32
export LFRING_BENCH_THREAD_NUM=12
producers=(1 2 3 4 6 8 9 10 11)
for p in "${producers[@]}"
do
   export LFRING_BENCH_PRODUCER_NUM=$p
   do_bench "$p" $LFRING_BENCH_CHARTS_FILE
done

echo \#end >> $LFRING_BENCH_CHARTS_FILE

### generate reports
go test -run "^TestGenReport$"