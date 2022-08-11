package lfring

import (
	. "gopkg.in/check.v1"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
)

func (s *MySuite) TestNodeMpmcConcurrencyRW(c *C) {
	concurrencyRW(c, NodeBased, func(buffer RingBuffer[*string]) uint64 {
		return atomic.LoadUint64(&buffer.(*nodeBased[*string]).head)
	})
}

func (s *MySuite) TestHybridMpmcConcurrencyRW(c *C) {
	concurrencyRW(c, Classical, func(buffer RingBuffer[*string]) uint64 {
		return atomic.LoadUint64(&buffer.(*classical[*string]).head)
	})
}

func (s *MySuite) TestMPSCConcurrencyRW(c *C) {
	// given
	source := initDataSource()

	capacity := 4
	buffer := New[*string](Classical, uint64(capacity)).(*classical[*string])

	var wg sync.WaitGroup
	offerNumber := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerAlphabet := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+8]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerPunctuation := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+16]
			for !buffer.Offer(&v) {
			}
		}
	}

	resultArr := make([]*string, 24)
	var finishWg sync.WaitGroup
	consumer := func(buffer RingBuffer[*string], ch chan struct{}) {
		counter := 0
		finishWg.Add(1)
		for {
			select {
			case <-ch:
				finishWg.Done()
				return
			default:
				buffer.SingleConsumerPoll(func(v *string) {
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
	for atomic.LoadUint64(&buffer.head) < 24 {
		runtime.Gosched()
	}
	close(done)
	finishWg.Wait()

	// then
	countSet := make(map[*string]int)
	drainToSet(resultArr, countSet)
	if len(countSet) != 24 {
		c.Assert(len(countSet), Equals, 24)
	}
	c.Assert(len(countSet), Equals, 24)
	for _, v := range countSet {
		c.Assert(v, Equals, 1)
	}
}

func (s *MySuite) TestMPSCVecConcurrencyRW(c *C) {
	// given
	source := initDataSource()

	capacity := 4
	buffer := New[*string](Classical, uint64(capacity)).(*classical[*string])

	var wg sync.WaitGroup
	offerNumber := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerAlphabet := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+8]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerPunctuation := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+16]
			for !buffer.Offer(&v) {
			}
		}
	}

	resultArr := make([]*string, 24)
	var finishWg sync.WaitGroup
	consumer := func(buffer RingBuffer[*string], ch chan struct{}) {
		counter := 0
		ret := make([]*string, capacity)
		finishWg.Add(1)
		for {
			select {
			case <-ch:
				finishWg.Done()
				return
			default:
				validCnt := buffer.SingleConsumerPollVec(ret)
				for i := uint64(0); i < validCnt; i++ {
					resultArr[counter] = ret[i]
					counter++
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
	for atomic.LoadUint64(&buffer.head) < 24 {
		runtime.Gosched()
	}

	close(done)
	finishWg.Wait()

	// then
	countSet := make(map[*string]int)
	drainToSet(resultArr, countSet)
	if len(countSet) != 24 {
		c.Assert(len(countSet), Equals, 24)
	}
	c.Assert(len(countSet), Equals, 24)
	for _, v := range countSet {
		c.Assert(v, Equals, 1)
	}
}

func (s *MySuite) TestSPMCConcurrencyRW(c *C) {

	// given
	source := initDataSource()

	capacity := 4
	buffer := New[*string](Classical, uint64(capacity)).(*classical[*string])

	var wg sync.WaitGroup
	producer := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		i := 0
		for {
			buffer.SingleProducerOffer(func() (v *string, finish bool) {
				if i == len(source) {
					return nil, true
				}
				v = &source[i]
				i++
				return
			})

			if i == 24 {
				break
			}
		}
	}

	var finishWg sync.WaitGroup
	consumer := func(buffer RingBuffer[*string], ch chan struct{}, outputArr []*string) {
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
	resultArr1 := make([]*string, 24)
	resultArr2 := make([]*string, 24)
	resultArr3 := make([]*string, 24)
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
	countSet := make(map[*string]int)
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

func concurrencyRW(c *C, t BufferType, getHead func(buffer RingBuffer[*string]) uint64) {
	// given
	source := initDataSource()

	capacity := 4
	buffer := New[*string](t, uint64(capacity))

	var wg sync.WaitGroup
	offerNumber := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerAlphabet := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+8]
			for !buffer.Offer(&v) {
			}
		}
	}

	offerPunctuation := func(buffer RingBuffer[*string]) {
		defer wg.Done()
		for i := 0; i < 8; i++ {
			v := source[i+16]
			for !buffer.Offer(&v) {
			}
		}
	}

	var finishWg sync.WaitGroup
	consumer := func(buffer RingBuffer[*string], ch chan struct{}, outputArr []*string) {
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
	resultArr1 := make([]*string, 24)
	resultArr2 := make([]*string, 24)
	resultArr3 := make([]*string, 24)
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
	countSet := make(map[*string]int)
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

func drainToSet(srcArr []*string, descSet map[*string]int) {
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
