package go_lock_free_ring_buffer

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkNodeMPMC(b *testing.B) {
	mpmcRB := New(NodeBasedMPMC, 16)
	mpmc(b, mpmcRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) / 2)
}

func BenchmarkHybridMPMC(b *testing.B) {
	mpscRB := New(Hybrid, 16)
	mpmc(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) / 2)
}

func BenchmarkChannelMPMC(b *testing.B) {
	fakeB := newFakeBuffer(16)
	mpmc(b, fakeB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) / 2)
}

func BenchmarkHybridMPSCControl(b *testing.B) {
	mpscRB := New(Hybrid, 16)
	mpmc(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) - 1)
}

func BenchmarkHybridMPSC(b *testing.B) {
	mpscRB := New(Hybrid, 16)
	mpsc(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) - 1)
}

func setup() []int {
	ints := make([]int, 64)
	for i := 0; i < len(ints); i++ {
		ints[i] = rand.Int()
	}

	return ints
}

type fakeBuffer struct {
	capacity uint64
	ch chan interface{}
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

func (r *fakeBuffer) SingleProducerOffer(valueSupplier func() (v interface{}, finish bool))  {
	v, finish := valueSupplier()
	if finish {
		return
	}

	r.ch <- v
}

func (r *fakeBuffer) SingleConsumerPoll(valueConsumer func(interface{}))  {
	v := <-r.ch
	valueConsumer(v)
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

func mpmc(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
	baseBenchmark(b, buffer, threadCount, trueCount,
		func(b RingBuffer, v interface{}) {
			b.Offer(v)
		},
		func(b RingBuffer, counter *int32) {
			if _, success := b.Poll(); success {
				atomic.AddInt32(counter, 1)
			}
		})
}

func mpsc(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
	baseBenchmark(b, buffer, threadCount, trueCount,
		func(b RingBuffer, v interface{}) {
			b.Offer(v)
		},
		func(b RingBuffer, counter *int32) {
			b.SingleConsumerPoll(func(v interface{}) {
				atomic.AddInt32(counter, 1)
			})
		})
}

func baseBenchmark(b *testing.B, buffer RingBuffer, threadCount int, trueCount int, offer func(b RingBuffer, v interface{}), poll func(b RingBuffer, counter *int32)) {
	ints := setup()

	counter := int32(0)
	go manage(b, threadCount, trueCount)
	b.RunParallel(func(pb *testing.PB) {
		producer := <-controlCh
		wg.Wait()
		for i := 1; pb.Next(); i++ {
			if producer {
				offer(buffer, ints[(i & (len(ints) - 1))])
			} else {
				poll(buffer, &counter)
			}
		}
	})

	b.StopTimer()
	b.Logf("Success handover count: %d", counter)
}
