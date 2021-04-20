package go_mpsc_ring_buffer

import (
	"sync/atomic"
	"unsafe"
)

type hybrid struct {
	ringBufferBasement
	element []interface{}
}

func newHybrid(capacity uint64) RingBuffer {
	return &hybrid{
		ringBufferBasement{head: uint64(0),
			tail: uint64(0),
			capacity: capacity,
			mask: capacity - 1,
		},
		make([]interface{}, capacity),
	}
}

// Offer a value pointer.
func (r *hybrid) Offer(value interface{}) (success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return false
	}

	newTail := oldTail + 1
	tailNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail & r.mask])))
	// not published yet
	if tailNode != nil {
		return false
	}
	if !atomic.CompareAndSwapUint64(&r.tail, oldTail, newTail) {
		return false
	}

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail & r.mask])), unsafe.Pointer(&value))
	return true
}

// Poll head value pointer.
func (r *hybrid) Poll() (value interface{}, success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isEmpty(oldTail, oldHead) {
		return nil, false
	}

	newHead := oldHead + 1
	headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead & r.mask])))
	// not published yet
	if headNode == nil {
		return nil, false
	}
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return nil, true
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead & r.mask])), nil)

	return *(*interface{})(headNode), true
}

// isFull check whether buffer is full by compare (tail - head).
// Because of none-sync read of tail and head, the tail maybe smaller than head(which is
// never happened in the view of buffer):
//
// Say if the thread read tail=4 at time point one (in this time head=3), then wait to
// get scheduled, after a long wait, at time point two (in this time tail=8), the thread
// read head=7. So at the view in the thread, tail=4 and head=7.
//
// Hence, once tail < head means the tail is far behind the real (which means CAS-tail will
// definitely fail), so we just return full to the Offer caller let it try again.
func (r *hybrid) isFull(tail uint64, head uint64) bool {
	return tail - head >= r.capacity - 1
}

// isEmpty check whether buffer is empty by compare (tail - head).
// Same as isFull, the tail also may be smaller than head at thread view, which can be lead
// to wrong result:
//
// consider consumer c1 get tail=3 head=5, but actually the latest tail=5
// (means buffer empty), if we continue Poll, the dirty value can be fetched, maybe nil -
// which is harmless, maybe old value that have not been set to nil yet (by last consumer)
// - which is fatal.
//
// To keep the correctness of ring buffer, we need to return true when tail < head and
// tail == head.
func (r *hybrid) isEmpty(tail uint64, head uint64) bool {
	return (tail < head) || (tail - head == 0)
}