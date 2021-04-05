package go_mpsc_ring_buffer

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestFindPowerOfTwo(c *C) {
	// given
	cap1 := uint64(0)
	cap2 := uint64(10)
	cap3 := uint64(16)
	cap4 := ^uint64(0)

	// when
	res1 := findPowerOfTwo(cap1)
	res2 := findPowerOfTwo(cap2)
	res3 := findPowerOfTwo(cap3)
	res4 := findPowerOfTwo(cap4)

	// then
	c.Assert(res1, Equals, uint64(0))
	c.Assert(res2, Equals, uint64(16))
	c.Assert(res3, Equals, uint64(16))
	c.Assert(res4, Equals, uint64(1<<63))
}

func (s *MySuite) TestOfferAndPollSuccess(c *C) {
	// given
	fakeString := "fake"
	buffer := New(10)

	// when
	result := buffer.Offer(&fakeString)
	poll, _ := buffer.Poll()

	// then
	c.Assert(result, Equals, true)
	c.Assert(poll, Equals, &fakeString)
}

func (s *MySuite) TestOfferFailedWhenFull(c *C) {
	// given
	capacity := 10
	buffer := New(uint64(capacity)).(*Mpsc)
	realCapacity := findPowerOfTwo(uint64(capacity + 1))
	for i := 0; i < int(realCapacity); i++ {
		buffer.Offer(i)
	}
	c.Assert(buffer.isFull(buffer.tail, buffer.head), Equals, true)
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(15))

	// when
	offered := buffer.Offer(10)

	// then
	c.Assert(offered, Equals, false)
}

func (s *MySuite) TestPollFailedWhenEmpty(c *C) {
	// given
	capacity := 10
	buffer := New(uint64(capacity)).(*Mpsc)
	c.Assert(buffer.isEmpty(buffer.tail, buffer.head), Equals, true)
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(0))

	// when
	_, empty := buffer.Poll()

	// then
	c.Assert(empty, Equals, true)
}

func (s *MySuite) TestRingBufferShift(c *C) {
	// given
	capacity := 10
	buffer := New(uint64(capacity)).(*Mpsc)

	// when
	for i := 0; i < 13; i++ {
		buffer.Offer(i)
	}

	// then
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(13))

	// when
	buffer.Offer(14)
	buffer.Offer(15)

	// then
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(15))
	c.Assert(buffer.isFull(buffer.tail, buffer.head), Equals, true)

	// when
	buffer.Poll()
	buffer.Offer(16)

	// then
	c.Assert(buffer.head, Equals, uint64(1))
	c.Assert(buffer.tail, Equals, uint64(0))
	c.Assert(buffer.isFull(buffer.tail, buffer.head), Equals, true)

	// when
	for i := 0; i < 14; i++ {
		buffer.Poll()
	}
	buffer.Offer(12)
	buffer.Offer(13)
	buffer.Offer(14)
	buffer.Poll()
	buffer.Poll()

	// then
	c.Assert(buffer.head, Equals, uint64(1))
	c.Assert(buffer.tail, Equals, uint64(3))
}