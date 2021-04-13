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
	realCapacity := findPowerOfTwo(capacity)
	return &Mpsc{
		make([]interface{}, realCapacity),
		0,
		0,
		realCapacity,
		realCapacity - 1,
	}
}

// findPowerOfTwo return the input number as round up to it's power of two
// The algorithm only care about the MSB of (givenNum -1), through the below procedure,
// the MSB will be spread to all lower bit than MSB. At last do (givenNum + 1) we
// can get power of two form of givenNum.
func findPowerOfTwo(givenMum uint64) uint64 {
	givenMum--
	givenMum |= givenMum >> 1
	givenMum |= givenMum >> 2
	givenMum |= givenMum >> 4
	givenMum |= givenMum >> 8
	givenMum |= givenMum >> 16
	givenMum |= givenMum >> 32
	givenMum++

	return givenMum
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