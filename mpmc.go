package go_mpsc_ring_buffer

type mpmc struct {
	element []interface{}
	head uint64
	tail uint64
	capacity uint64
	mask uint64
}

// Offer a value pointer.
func (r *mpmc) Offer(valuePointer interface{}) bool {
	return false
}

// Poll head value pointer.
func (r *mpmc) Poll() (valuePointer interface{}, empty bool) {
	return nil, false
}

func (r *mpmc) isEmpty(tail uint64, head uint64) bool {
	return false
}

func (r *mpmc) isFull(tail uint64, head uint64) bool {
	return false
}