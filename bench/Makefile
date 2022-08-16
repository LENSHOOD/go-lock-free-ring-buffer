mpmc-benchmark:
	./run-bench-mpmc.sh

mpmc-cpu-profile:
	env LFRING_BENCH_THREAD_NUM=12 LFRING_BENCH_PRODUCER_NUM=6 LFRING_BENCH_CAP=32 go test -run "^$$" -bench "^.+(NodeMPMC|HybridMPMC)$$" -benchtime=10s -count=10 -cpuprofile cpuprofile.out

gen-report:
ifeq ($(LFRING_BENCH_CHARTS_FILE),)
	$(error Please set env LFRING_BENCH_CHARTS_FILE as the dat file ready to generate report)
else
	env LFRING_BENCH_THREAD_NUM=12 LFRING_BENCH_PRODUCER_NUM=6 LFRING_BENCH_CAP=32 go test -run "^TestGenReport$$"
endif

clean:
	rm -f *.html