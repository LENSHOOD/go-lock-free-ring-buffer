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
	poll := buffer.Poll()

	// then
	c.Assert(result, Equals, true)
	c.Assert(poll, Equals, fakeString)
}