package go_mpsc_ring_buffer

import (
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func BenchmarkNodeMPMC(b *testing.B) {
	mpmcRB := New(MPMC, 16)
	baseBenchmark(b, mpmcRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) / 2)
}

func BenchmarkHybridMPMC(b *testing.B) {
	mpscRB := New(Hybrid, 16)
	baseBenchmark(b, mpscRB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) - 1)
}

func BenchmarkChannelMPMC(b *testing.B) {
	fakeB := newFakeBuffer(16)
	baseBenchmark(b, fakeB, runtime.GOMAXPROCS(0), runtime.GOMAXPROCS(0) / 2)
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

func baseBenchmark(b *testing.B, buffer RingBuffer, threadCount int, trueCount int) {
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
