package go_mpsc_ring_buffer

import (
	. "gopkg.in/check.v1"
	"testing"
)

// hook up go-check to go testing
func Test(t *testing.T) {TestingT(t)}

type MySuite struct {}

var _ = Suite(&MySuite{})

func (s *MySuite) TestOfferAndPollSuccess(c *C) {
	// given
	fakeString := "fake"
	buffer := New(10)

	// when
	result := buffer.Offer(fakeString)
	poll, _ := buffer.Poll()

	// then
	c.Assert(result, Equals, true)
	c.Assert(poll, Equals, fakeString)
}

func (s *MySuite) TestOfferFailedWhenFull(c *C) {
	// given
	capacity := 10
	buffer := New(uint64(capacity)).(*Mpsc)
	for i := 0; i < capacity; i++ {
		buffer.Offer(i)
	}
	c.Assert(buffer.isFull(), Equals, true)
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(10))

	// when
	offered := buffer.Offer(10)

	// then
	c.Assert(offered, Equals, false)
}

func (s *MySuite) TestPollFailedWhenEmpty(c *C) {
	// given
	capacity := 10
	buffer := New(uint64(capacity)).(*Mpsc)
	c.Assert(buffer.isEmpty(), Equals, true)
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(1))

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
	for i := 0; i < 8; i++ {
		buffer.Offer(i)
	}

	// then
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(9))

	// when
	buffer.Offer(9)
	buffer.Offer(10)

	// then
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(10))
	c.Assert(buffer.isFull(), Equals, true)

	// when
	buffer.Poll()
	buffer.Offer(11)

	// then
	c.Assert(buffer.head, Equals, uint64(1))
	c.Assert(buffer.tail, Equals, uint64(0))
	c.Assert(buffer.isFull(), Equals, true)

	// when
	for i := 0; i < 8; i++ {
		buffer.Poll()
	}
	buffer.Offer(12)
	buffer.Offer(13)
	buffer.Offer(14)
	buffer.Poll()
	buffer.Poll()

	// then
	c.Assert(buffer.head, Equals, uint64(0))
	c.Assert(buffer.tail, Equals, uint64(3))
}