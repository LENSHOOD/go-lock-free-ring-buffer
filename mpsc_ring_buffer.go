package go_mpsc_ring_buffer

type OfferStatus int

type MpscRingBuffer interface {
	Offer(interface{}) bool
	Poll() (value interface{}, empty bool)
}

type Mpsc struct {
	element []interface{}
	head uint64
	tail uint64
	capacity uint64
	mask uint64
}

// New MpscRingBuffer with Mpsc.
// the array capacity should add extra one because ring buffer always leave one slot empty
// expand capacity as power-of-two, to make head/tail calculate faster and simpler
func New(capacity uint64) MpscRingBuffer {
	realCapacity := findPowerOfTwo(capacity + 1)
	return &Mpsc{
		make([]interface{}, realCapacity),
		0,
		1,
		realCapacity,
		realCapacity - 1,
	}
}

func findPowerOfTwo(givenMum uint64) uint64 {
	if givenMum == 0 {
		return 0
	}

	var bottom uint64 = 0
	var top uint64 = 63
	for {
		if top-bottom <= 1 {
			return 1 << top
		}

		mid := (top - bottom) / 2
		if givenMum > (1 << (bottom + mid)) {
			bottom = bottom + mid
		} else {
			top = bottom + mid
		}
	}
}

func (r *Mpsc) Offer(node interface{}) bool {
	if r.isFull() {
		return false
	}

	r.element[r.tail] = node
	r.tail = (r.tail+1) & r.mask

	return true
}

func (r *Mpsc) Poll() (value interface{}, empty bool) {
	if r.isEmpty() {
		return nil, true
	}

	r.head = (r.head+1) & r.mask
	headNode := r.element[r.head]

	return headNode, false
}

func (r *Mpsc) isEmpty() bool {
	return (r.tail - r.head) & r.mask == 1
}

func (r *Mpsc) isFull() bool {
	return (r.head - r.tail)  & r.mask == 0
}