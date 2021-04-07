package go_mpsc_ring_buffer

import (
	. "gopkg.in/check.v1"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

func (s *MySuite) TestConcurrencyRW(c *C) {
	// given
	capacity := 4
	buffer := New(uint64(capacity)).(*Mpsc)

	var wg sync.WaitGroup

	offerNumber := func(buffer MpscRingBuffer) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := strconv.Itoa(i)
			for !buffer.Offer(&v) {}
		}
	}

	offerAlphabet := func(buffer MpscRingBuffer) {
		defer wg.Done()
		asciiStart := 65
		for i := 0; i < 8; i++ {
			v := string(rune(asciiStart))
			asciiStart += 1
			for !buffer.Offer(&v) {}
		}
	}

	offerPunctuation := func(buffer MpscRingBuffer) {
		defer wg.Done()
		asciiStart := 33
		for i := 0; i < 8; i++ {
			v := string(rune(asciiStart))
			asciiStart += 1
			for !buffer.Offer(&v) {}
		}
	}

	resultSet := make(map[interface{}]bool)
	var finishWg sync.WaitGroup
	consumer := func(buffer MpscRingBuffer, ch chan struct{}) {
		finishWg.Add(1)
		for  {
			select {
			case <- ch:
				finishWg.Done()
				return
			default:
				if poll, empty := buffer.Poll(); !empty {
					resultSet[*(poll.(*string))] = true
				}
			}
		}
	}
	// when
	done := make(chan struct{})
	wg.Add(3)
	go consumer(buffer, done)
	go offerNumber(buffer)
	go offerAlphabet(buffer)
	go offerPunctuation(buffer)

	wg.Wait()
	for !buffer.isEmpty(atomic.LoadUint64(&buffer.tail), atomic.LoadUint64(&buffer.head)) {
		runtime.Gosched()
	}
	close(done)
	finishWg.Wait()

	// then
	c.Assert(len(resultSet), Equals, 24)
}