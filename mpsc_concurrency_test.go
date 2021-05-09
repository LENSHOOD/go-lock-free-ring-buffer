package lfring

import (
	. "gopkg.in/check.v1"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

func (s *MySuite) TestMPSCConcurrencyRW(c *C) {
	// given
	source := initDataSource()

	capacity := 4
	buffer := New(Classical, uint64(capacity)).(*classical)

	var wg sync.WaitGroup
	offerNumber := func(buffer RingBuffer) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerAlphabet := func(buffer RingBuffer) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+8]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerPunctuation := func(buffer RingBuffer) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+16]
			for !buffer.Offer(&v) {
			}
		}
	}

	resultArr := make([]interface{}, 24)
	var finishWg sync.WaitGroup
	consumer := func(buffer RingBuffer, ch chan struct{}) {
		counter := 0
		finishWg.Add(1)
		for {
			select {
			case <-ch:
				finishWg.Done()
				return
			default:
				buffer.SingleConsumerPoll(func(v interface{}) {
					resultArr[counter] = v
					counter++
				})
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
	countSet := make(map[interface{}]bool)
	for i := 0; i < len(resultArr); i++ {
		if resultArr[i] != nil {
			countSet[resultArr[i]] = true
		}
	}
	c.Assert(len(countSet), Equals, 24)
}

func initDataSource() []string {
	sourceArray := make([]string, 24)
	for i := 0; i < 24; i++ {
		var v string
		if i < 8 {
			v = strconv.Itoa(i)
		} else if i >= 8 && i < 16 {
			v = string(rune(65 + i - 8))
		} else {
			v = string(rune(33 + i - 16))
		}
		sourceArray[i] = v
	}
	return sourceArray
}
