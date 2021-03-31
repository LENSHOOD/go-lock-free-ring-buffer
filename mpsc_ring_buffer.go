package go_mpsc_ring_buffer

type OfferStatus int

type MpscRingBuffer interface {
	Offer(interface{}) bool
	Poll() interface{}
}

type Mpsc struct {
	element []interface{}
	head uint64
	tail uint64
	capacity uint64
}

func New(capacity uint64) MpscRingBuffer {
	return &Mpsc{
		make([]interface{}, capacity),
		0,
		1,
		capacity,
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

func (r *Mpsc) Poll() interface{} {
	if r.isEmpty() {
		return nil
	}

	headNode := r.element[r.head + 1]

	r.head++
	if r.head == r.capacity {
		r.head = 0
	}

	return headNode
}

func (r *Mpsc) isEmpty() bool {
	return (r.tail - r.head) == 1 || (r.head - r.tail) == r.capacity - 1
}

func (r *Mpsc) isFull() bool {
	return (r.head - r.tail) == 1 || (r.tail - r.head) == r.capacity - 1
}