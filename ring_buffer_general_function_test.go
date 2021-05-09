package lfring

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestFindPowerOfTwo(c *C) {
	// given
	cap1 := uint64(0)
	cap2 := uint64(10)
	cap3 := uint64(16)
	cap4 := uint64(33)
	cap5 := ^uint64(0)

	// when
	res1 := findPowerOfTwo(cap1)
	res2 := findPowerOfTwo(cap2)
	res3 := findPowerOfTwo(cap3)
	res4 := findPowerOfTwo(cap4)
	res5 := findPowerOfTwo(cap5)

	// then
	c.Assert(res1, Equals, uint64(0))
	c.Assert(res2, Equals, uint64(16))
	c.Assert(res3, Equals, uint64(16))
	c.Assert(res4, Equals, uint64(64))
	c.Assert(res5, Equals, uint64(0))
}

var bufferSet = []BufferType{NodeBased, Classical}

func (s *MySuite) TestOfferAndPollSuccess(c *C) {
	for _, t := range bufferSet {
		// given
		fakeString := "fake"
		buffer := New(t, 10)

		// when
		result := buffer.Offer(&fakeString)
		poll, _ := buffer.Poll()

		// then
		c.Assert(result, Equals, true)
		c.Assert(poll, Equals, &fakeString)
	}
}

func (s *MySuite) TestOfferFailedWhenFull(c *C) {
	for _, t := range bufferSet {
		// given
		capacity := 10
		buffer := New(t, uint64(capacity))
		realCapacity := findPowerOfTwo(uint64(capacity + 1))
		for i := 0; i < int(realCapacity); i++ {
			buffer.Offer(i)
		}

		// when
		offered := buffer.Offer(10)

		// then
		c.Assert(offered, Equals, false)
	}
}

func (s *MySuite) TestPollFailedWhenEmpty(c *C) {
	for _, t := range bufferSet {
		// given
		capacity := 10
		buffer := New(t, uint64(capacity))

		// when
		_, success := buffer.Poll()

		// then
		c.Assert(success, Equals, false)
	}
}

func (s *MySuite) TestRingBufferShift(c *C) {
	for _, t := range bufferSet {
		// given
		capacity := 10
		buffer := New(t, uint64(capacity))

		// when
		for i := 0; i < 13; i++ {
			buffer.Offer(i)
		}

		// when
		buffer.Offer(13)
		buffer.Offer(14)

		// then
		polled, success := buffer.Poll()
		c.Assert(success, Equals, true)
		c.Assert(polled, Equals, 0)

		// when
		buffer.Offer(15)

		// then
		for i := 0; i < 14; i++ {
			polled, success := buffer.Poll()
			c.Assert(success, Equals, true)
			c.Assert(polled, Equals, i+1)
		}

		// when
		buffer.Offer(16)
		buffer.Offer(17)
		buffer.Offer(18)

		// then
		polled1, _ := buffer.Poll()
		c.Assert(polled1, Equals, 15)
		polled2, _ := buffer.Poll()
		c.Assert(polled2, Equals, 16)
	}
}
