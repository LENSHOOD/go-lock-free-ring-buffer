package go_mpsc_ring_buffer

import (
	"sync/atomic"
	"unsafe"
)

type Mpsc struct {
	element []interface{}
	head uint64
	tail uint64
	capacity uint64
	mask uint64
}

// Offer a value pointer.
func (r *Mpsc) Offer(valuePointer interface{}) bool {
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

	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newTail & r.mask])), unsafe.Pointer(&valuePointer))
	return true
}

// Poll head value pointer.
func (r *Mpsc) Poll() (valuePointer interface{}, empty bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isEmpty(oldTail, oldHead) {
		return nil, true
	}

	newHead := oldHead + 1
	headNode := atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead & r.mask])))
	// not published yet
	if headNode == nil {
		return nil, true
	}
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return nil, true
	}
	atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&r.element[newHead & r.mask])), nil)

	return *(*interface{})(headNode), false
}

func (r *Mpsc) isEmpty(tail uint64, head uint64) bool {
	return tail - head == 0
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
func (r *Mpsc) isFull(tail uint64, head uint64) bool {
	return tail - head >= r.capacity - 1
}