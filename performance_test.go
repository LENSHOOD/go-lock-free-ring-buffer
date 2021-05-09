package lfring

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkNodeMPMC(b *testing.B) {
	mpmcRB := New(NodeBased, 16)
	mpmcBenchmark(b, mpmcRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0)/2)
}

func BenchmarkHybridMPMC(b *testing.B) {
	mpscRB := New(Classical, 16)
	mpmcBenchmark(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0)/2)
}

func BenchmarkChannelMPMC(b *testing.B) {
	fakeB := newFakeBuffer(16)
	mpmcBenchmark(b, fakeB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0)/2)
}

func BenchmarkHybridMPSCControl(b *testing.B) {
	mpscRB := New(Classical, 16)
	mpmcBenchmark(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0)-1)
}

func BenchmarkHybridMPSC(b *testing.B) {
	mpscRB := New(Classical, 16)
	mpscBenchmark(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0)-1)
}

func BenchmarkHybridSPMCControl(b *testing.B) {
	mpscRB := New(Classical, 16)
	mpmcBenchmark(b, mpscRB, runtime.GOMAXPROCS(0), 1)
}

func BenchmarkHybridSPMC(b *testing.B) {
	mpscRB := New(Classical, 16)
	spmcBenchmark(b, mpscRB, runtime.GOMAXPROCS(0), 1)
}

func BenchmarkHybridSPSCControl(b *testing.B) {
	runtime.GOMAXPROCS(2)
	mpscRB := New(Classical, 16)
	mpmcBenchmark(b, mpscRB, 2, 1)
}

func BenchmarkHybridSPSC(b *testing.B) {
	runtime.GOMAXPROCS(2)
	mpscRB := New(Classical, 16)
	spscBenchmark(b, mpscRB, 2, 1)
}

type fakeBuffer struct {
	capacity uint64
	ch       chan interface{}
}

func newFakeBuffer(capacity uint64) RingBuffer {
	return &fakeBuffer{
		capacity,
		make(chan interface{}, capacity),
	}
}

func (r *fakeBuffer) Offer(value interface{}) (success bool) {
	select {
	case r.ch <- value:
		return true
	default:
		return false
	}
}

func (r *fakeBuffer) Poll() (value interface{}, success bool) {
	select {
	case v := <-r.ch:
		return v, true
	default:
		return nil, false
	}
}

func (r *fakeBuffer) SingleProducerOffer(valueSupplier func() (v interface{}, finish bool)) {
	v, finish := valueSupplier()
	if finish {
		return
	}

	r.ch <- v
}

func (r *fakeBuffer) SingleConsumerPoll(valueConsumer func(interface{})) {
	v := <-r.ch
	valueConsumer(v)
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
}

func mpmcBenchmark(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	go manage(b, threadCount, trueCount)
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
	b.Logf("Success handover count: %d", counter)
}

func mpscBenchmark(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	consumer := func(v interface{}) {
		atomic.AddInt32(&counter, 1)
	}
	go manage(b, threadCount, trueCount)
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
	b.Logf("Success handover count: %d", counter)
}

func spmcBenchmark(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	go manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				j := i
				buffer.SingleProducerOffer(func() (v interface{}, finish bool) {
					v = ints[(j & (len(ints) - 1))]
					j++
					return v, false
				})
			} else {
				if _, success := buffer.Poll(); success {
					atomic.AddInt32(&counter, 1)
				}
			}
		}
	})

	b.StopTimer()
	b.Logf("Success handover count: %d", counter)
}

func spscBenchmark(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
	ints := setup()

	counter := int32(0)
	consumer := func(v interface{}) {
		atomic.AddInt32(&counter, 1)
	}
	go manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				j := i
				buffer.SingleProducerOffer(func() (v interface{}, finish bool) {
					v = ints[(j & (len(ints) - 1))]
					j++
					return v, false
				})
			} else {
				buffer.SingleConsumerPoll(consumer)
			}
		}
	})

	b.StopTimer()
	b.Logf("Success handover count: %d", counter)
}
