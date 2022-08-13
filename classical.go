package lfring

import (
	"sync/atomic"
)

type classical[T any] struct {
	head     uint64
	tail     uint64
	capacity uint64
	mask     uint64
	element  []*T
}

func newClassical[T any](capacity uint64) RingBuffer[T] {
	return &classical[T]{
		head:     uint64(0),
		tail:     uint64(0),
		capacity: capacity,
		mask:     capacity - 1,
		element:  make([]*T, capacity),
	}
}

func (r *classical[T]) Offer(value T) (success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return false
	}

	newTail := oldTail + 1
	tailNode := r.element[newTail&r.mask]
	// not published yet
	if tailNode != nil {
		return false
	}
	if !atomic.CompareAndSwapUint64(&r.tail, oldTail, newTail) {
		return false
	}

	r.element[newTail&r.mask] = &value
	return true
}

func (r *classical[T]) SingleProducerOffer(valueSupplier func() (v T, finish bool)) {
	oldTail := r.tail
	oldHead := atomic.LoadUint64(&r.head)
	if r.isFull(oldTail, oldHead) {
		return
	}

	newTail := oldTail + 1
	for ; newTail-oldHead < r.capacity; newTail++ {
		tailNode := r.element[newTail&r.mask]
		// not published yet
		if tailNode != nil {
			break
		}

		v, finish := valueSupplier()
		if finish {
			break
		}
		r.element[newTail&r.mask] = &v
	}
	atomic.StoreUint64(&r.tail, newTail-1)
}

func (r *classical[T]) Poll() (value T, success bool) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := atomic.LoadUint64(&r.head)
	if r.isEmpty(oldTail, oldHead) {
		return
	}

	newHead := oldHead + 1
	headNode := r.element[newHead&r.mask]
	// not published yet
	if headNode == nil {
		return
	}
	if !atomic.CompareAndSwapUint64(&r.head, oldHead, newHead) {
		return
	}
	r.element[newHead&r.mask] = nil

	return *headNode, true
}

func (r *classical[T]) SingleConsumerPoll(valueConsumer func(T)) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := r.head
	if r.isEmpty(oldTail, oldHead) {
		return
	}

	currHead := oldHead + 1
	for ; currHead <= oldTail; currHead++ {
		currNode := r.element[currHead&r.mask]
		// not published yet
		if currNode == nil {
			break
		}
		valueConsumer(*currNode)
		r.element[currHead&r.mask] = nil
	}

	atomic.StoreUint64(&r.head, currHead-1)
}

func (r *classical[T]) SingleConsumerPollVec(ret []T) (validCnt uint64) {
	oldTail := atomic.LoadUint64(&r.tail)
	oldHead := r.head
	if r.isEmpty(oldTail, oldHead) {
		return
	}

	currHead := oldHead + 1
	for ; currHead <= oldTail; currHead++ {
		currNode := r.element[currHead&r.mask]
		// not published yet
		if currNode == nil {
			break
		}
		ret[currHead-oldHead-1] = *currNode
		r.element[currHead&r.mask] = nil
	}

	atomic.StoreUint64(&r.head, currHead-1)

	return currHead - oldHead - 1
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
func (r *classical[T]) isFull(tail uint64, head uint64) bool {
	return tail-head >= r.capacity-1
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
func (r *classical[T]) isEmpty(tail uint64, head uint64) bool {
	return (tail < head) || (tail-head == 0)
}
