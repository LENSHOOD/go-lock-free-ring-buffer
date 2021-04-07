package go_mpsc_ring_buffer

import (
	"sync/atomic"
	"unsafe"
)

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
		0,
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

func (r *Mpsc) isFull(tail uint64, head uint64) bool {
	return tail - head >= r.capacity - 1
}