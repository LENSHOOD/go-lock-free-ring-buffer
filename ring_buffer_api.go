package go_mpsc_ring_buffer

type RingBuffer interface {
	Offer(interface{}) (success bool)
	Poll() (value interface{}, success bool)
}

type BufferType int
const (
	MPMC BufferType = iota
	MPSC
	SPMC
	SPSC
)

// New RingBuffer with BufferType.
// The array capacity should add extra one because ring buffer always left one slot empty.
// Expand capacity as power-of-two, to make head/tail calculate faster and simpler
func New(t BufferType, capacity uint64) RingBuffer {
	realCapacity := findPowerOfTwo(capacity)

	switch t {
	case MPMC:
		return newMpmc(realCapacity)
	case MPSC:
		return newMpsc(realCapacity)
	default:
		panic("shouldn't goes here.")
	}
}

