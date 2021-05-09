package lfring

import (
	. "gopkg.in/check.v1"
	"runtime"
	"sync"
	"sync/atomic"
)

func (s *MySuite) TestNodeMpmcConcurrencyRW(c *C) {
	concurrencyRW(c, NodeBased, func(buffer RingBuffer) uint64 {
		return atomic.LoadUint64(&buffer.(*nodeBased).head)
	})
}

func (s *MySuite) TestHybridMpmcConcurrencyRW(c *C) {
	concurrencyRW(c, Classical, func(buffer RingBuffer) uint64 {
		return atomic.LoadUint64(&buffer.(*classical).head)
	})
}

func concurrencyRW(c *C, t BufferType, getHead func(buffer RingBuffer) uint64) {
	// given
	source := initDataSource()

	capacity := 4
	buffer := New(t, uint64(capacity))

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

	wg.Add(3)
	go offerNumber(buffer)
	go offerAlphabet(buffer)
	go offerPunctuation(buffer)

	wg.Wait()
	for getHead(buffer) < 24 {
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

func drainToSet(srcArr []interface{}, descSet map[interface{}]int) {
	for i := 0; i < len(srcArr); i++ {
		if srcArr[i] != nil {
			if v, exist := descSet[srcArr[i]]; exist {
				descSet[srcArr[i]] = v + 1
			} else {
				descSet[srcArr[i]] = 1
			}
		}
	}
}
