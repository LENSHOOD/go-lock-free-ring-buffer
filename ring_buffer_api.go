package go_mpsc_ring_buffer

type RingBuffer interface {
	Offer(interface{}) bool
	Poll() (value interface{}, empty bool)
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
func New(t BufferType, capacity uint64) MpscRingBuffer {
	realCapacity := findPowerOfTwo(capacity)

	switch t {
	case MPSC:
		return &Mpsc{
			make([]interface{}, realCapacity),
			0,
			0,
			realCapacity,
			realCapacity - 1,
		}
	default:
		panic("shouldn't goes here.")
	}
}

