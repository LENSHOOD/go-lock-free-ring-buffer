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
}

// New MpscRingBuffer with Mpsc.
// the array capacity should add extra one because ring buffer always leave one slot empty
func New(capacity uint64) MpscRingBuffer {
	return &Mpsc{
		make([]interface{}, capacity + 1),
		0,
		1,
		capacity + 1,
	}
}

func (r *Mpsc) Offer(node interface{}) bool {
	if r.isFull() {
		return false
	}

	r.element[r.tail] = node

	r.tail++
	if r.tail == r.capacity {
		r.tail = 0
	}

	return true
}

func (r *Mpsc) Poll() (value interface{}, empty bool) {
	if r.isEmpty() {
		return nil, true
	}

	r.head++
	if r.head == r.capacity {
		r.head = 0
	}

	headNode := r.element[r.head]

	return headNode, false
}

func (r *Mpsc) isEmpty() bool {
	return (r.tail - r.head) == 1 || (r.head - r.tail) == r.capacity - 1
}

func (r *Mpsc) isFull() bool {
	return (r.head - r.tail) == 1 || (r.tail - r.head) == r.capacity - 1
}