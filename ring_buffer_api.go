package lfring

// RingBuffer defines the behavior of ring buffer
type RingBuffer interface {
	Offer(interface{}) (success bool)
	Poll() (value interface{}, success bool)
	SingleProducerOffer(valueSupplier func() (v interface{}, finish bool))
	SingleConsumerPoll(valueConsumer func(interface{}))
}

// BufferType contains different type name of ring buffer
type BufferType int

const (
	// Hybrid ring buffer, hybrid with mpmc/mpsc/spmc
	Hybrid BufferType = iota

	// NodeBasedMPMC is a type of ring buffer that is node based
	NodeBasedMPMC
)

// New RingBuffer with BufferType.
// The array capacity should add extra one because ring buffer always left one slot empty.
// Expand capacity as power-of-two, to make head/tail calculate faster and simpler
func New(t BufferType, capacity uint64) RingBuffer {
	realCapacity := findPowerOfTwo(capacity)

	switch t {
	case NodeBasedMPMC:
		return newNodeBasedMpmc(realCapacity)
	case Hybrid:
		return newHybrid(realCapacity)
	default:
		panic("shouldn't goes here.")
	}
}

// findPowerOfTwo return the input number as round up to it's power of two
// The algorithm only care about the MSB of (givenNum -1), through the below procedure,
// the MSB will be spread to all lower bit than MSB. At last do (givenNum + 1) we
// can get power of two form of givenNum.
func findPowerOfTwo(givenMum uint64) uint64 {
	givenMum--
	givenMum |= givenMum >> 1
	givenMum |= givenMum >> 2
	givenMum |= givenMum >> 4
	givenMum |= givenMum >> 8
	givenMum |= givenMum >> 16
	givenMum |= givenMum >> 32
	givenMum++

	return givenMum
}
