package bench

import (
	"fmt"
	"github.com/LENSHOOD/go-lock-free-ring-buffer"
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

var (
	capacity        = toUint64(os.Getenv("LFRING_BENCH_CAP"))
	threadNum       = toInt(os.Getenv("LFRING_BENCH_THREAD_NUM"))
	mpmcProducerNum = toInt(os.Getenv("LFRING_BENCH_PRODUCER_NUM"))
)

func toInt(s string) (ret int) {
	return int(toUint64(s))
}

func toUint64(s string) (ret uint64) {
	ret, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		panic(fmt.Sprintf("wrong param: \"%s\", please check ENV", s))
	}

	return
}

func BenchmarkNodeMPMC(b *testing.B) {
	mpmcRB := lfring.New[int](lfring.NodeBased, capacity)
	mpmcBenchmark(b, mpmcRB, threadNum, mpmcProducerNum)
}

func BenchmarkHybridMPMC(b *testing.B) {
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	mpmcBenchmark(b, mpscRB, threadNum, mpmcProducerNum)
}

func BenchmarkChannelMPMC(b *testing.B) {
	fakeB := newFakeBuffer[int](capacity)
	mpmcBenchmark(b, fakeB, threadNum, mpmcProducerNum)
}

func BenchmarkHybridMPSCControl(b *testing.B) {
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	mpmcBenchmark(b, mpscRB, threadNum, threadNum-1)
}

func BenchmarkHybridMPSC(b *testing.B) {
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	mpscBenchmark(b, mpscRB, threadNum, threadNum-1)
}

func BenchmarkHybridMPSCVec(b *testing.B) {
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	mpscBenchmarkVec(b, mpscRB, threadNum, threadNum-1)
}

func BenchmarkHybridSPMCControl(b *testing.B) {
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	mpmcBenchmark(b, mpscRB, threadNum, 1)
}

func BenchmarkHybridSPMC(b *testing.B) {
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	spmcBenchmark(b, mpscRB, threadNum, 1)
}

func BenchmarkHybridSPSCControl(b *testing.B) {
	runtime.GOMAXPROCS(2)
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	mpmcBenchmark(b, mpscRB, 2, 1)
}

func BenchmarkHybridSPSC(b *testing.B) {
	runtime.GOMAXPROCS(2)
	mpscRB := lfring.New[int](lfring.Classical, capacity)
	spscBenchmark(b, mpscRB, 2, 1)
}

type fakeBuffer[T any] struct {
	capacity uint64
	ch       chan T
	empty    T
}

func newFakeBuffer[T any](capacity uint64) lfring.RingBuffer[T] {
	return &fakeBuffer[T]{
		capacity: capacity,
		ch:       make(chan T, capacity),
	}
}

func (r *fakeBuffer[T]) Offer(value T) (success bool) {
	select {
	case r.ch <- value:
		return true
	default:
		return false
	}
}

func (r *fakeBuffer[T]) Poll() (value T, success bool) {
	select {
	case v := <-r.ch:
		return v, true
	default:
		return r.empty, false
	}
}

func (r *fakeBuffer[T]) SingleProducerOffer(valueSupplier func() (v T, finish bool)) {
	v, finish := valueSupplier()
	if finish {
		return
	}

	r.ch <- v
}

func (r *fakeBuffer[T]) SingleConsumerPoll(valueConsumer func(T)) {
	v := <-r.ch
	valueConsumer(v)
}

func (r *fakeBuffer[T]) SingleConsumerPollVec(ret []T) (validCnt uint64) {
	return
}

func setup() []int {
	ints := make([]int, 64)
	for i := 0; i < len(ints); i++ {
		ints[i] = rand.Int()
	}

	return ints
}

var controlCh = make(chan bool)
var wg sync.WaitGroup

func manage(b *testing.B, threadCount int, trueCount int) {
	runtime.GOMAXPROCS(threadCount)

	go func() {
		wg.Add(1)
		for i := 0; i < threadCount; i++ {
			if trueCount > 0 {
				controlCh <- true
				trueCount--
			} else {
				controlCh <- false
			}
		}

		b.ResetTimer()
		wg.Done()
	}()
}

func mpmcBenchmark(b *testing.B, buffer lfring.RingBuffer[int], threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				buffer.Offer(ints[(i & (len(ints) - 1))])
			} else {
				if _, success := buffer.Poll(); success {
					atomic.AddInt32(&counter, 1)
				}
			}
		}
	})

	b.StopTimer()
	b.ReportMetric(float64(counter), "handovers")
}

func mpscBenchmark(b *testing.B, buffer lfring.RingBuffer[int], threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	consumer := func(v int) {
		atomic.AddInt32(&counter, 1)
	}
	manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				buffer.Offer(ints[(i & (len(ints) - 1))])
			} else {
				buffer.SingleConsumerPoll(consumer)
			}
		}
	})

	b.StopTimer()
	b.ReportMetric(float64(counter), "handovers")
}

func mpscBenchmarkVec(b *testing.B, buffer lfring.RingBuffer[int], threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	ret := make([]int, capacity, capacity)
	manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				buffer.Offer(ints[(i & (len(ints) - 1))])
			} else {
				validCnt := buffer.SingleConsumerPollVec(ret)
				atomic.AddInt32(&counter, int32(validCnt))
			}
		}
	})

	b.StopTimer()
	b.ReportMetric(float64(counter), "handovers")
}

func spmcBenchmark(b *testing.B, buffer lfring.RingBuffer[int], threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				j := i
				buffer.SingleProducerOffer(func() (v int, finish bool) {
					v = ints[(j & (len(ints) - 1))]
					j++
					return
				})
			} else {
				if _, success := buffer.Poll(); success {
					atomic.AddInt32(&counter, 1)
				}
			}
		}
	})

	b.StopTimer()
	b.ReportMetric(float64(counter), "handovers")
}

func spscBenchmark(b *testing.B, buffer lfring.RingBuffer[int], threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	consumer := func(v int) {
		atomic.AddInt32(&counter, 1)
	}
	manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				j := i
				buffer.SingleProducerOffer(func() (v int, finish bool) {
					v = ints[(j & (len(ints) - 1))]
					j++
					return
				})
			} else {
				buffer.SingleConsumerPoll(consumer)
			}
		}
	})

	b.StopTimer()
	b.ReportMetric(float64(counter), "handovers")
}
