package lfring

import (
	. "gopkg.in/check.v1"
	"runtime"
	"sync"
	"sync/atomic"
)

func (s *MySuite) TestSPMCConcurrencyRW(c *C) {

	// given
	source := initDataSource()

	capacity := 4
	buffer := New(Classical, uint64(capacity)).(*classical)

	var wg sync.WaitGroup
	producer := func(buffer RingBuffer) {
		defer wg.Done()
		i := 0
		for {
			buffer.SingleProducerOffer(func() (v interface{}, finish bool) {
				if i == len(source) {
					return nil, true
				}
				v = source[i]
				i++
				return &v, false
			})

			if i == 24 {
				break
			}
		}
	}

	var finishWg sync.WaitGroup
	consumer := func(buffer RingBuffer, ch chan struct{}, outputArr []interface{}) {
		counter := 0
		for {
			select {
			case <-ch:
				finishWg.Done()
				return
			default:
				if poll, success := buffer.Poll(); success {
					outputArr[counter] = poll
					counter++
				}
			}
		}
	}

	// when
	done := make(chan struct{})
	finishWg.Add(3)
	resultArr1 := make([]interface{}, 24)
	resultArr2 := make([]interface{}, 24)
	resultArr3 := make([]interface{}, 24)
	go consumer(buffer, done, resultArr1)
	go consumer(buffer, done, resultArr2)
	go consumer(buffer, done, resultArr3)

	wg.Add(1)
	go producer(buffer)

	wg.Wait()
	for atomic.LoadUint64(&buffer.head) < 24 {
		runtime.Gosched()
	}
	close(done)
	finishWg.Wait()

	// then
	countSet := make(map[interface{}]int)
	drainToSet(resultArr1, countSet)
	drainToSet(resultArr2, countSet)
	drainToSet(resultArr3, countSet)
	if len(countSet) != 24 {
		c.Assert(len(countSet), Equals, 24)
	}
	c.Assert(len(countSet), Equals, 24)
	for _, v := range countSet {
		c.Assert(v, Equals, 1)
	}
}
