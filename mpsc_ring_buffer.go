package go_mpsc_ring_buffer

type OfferStatus int

type MpscRingBuffer interface {
	offer(interface{}) bool
	poll() interface{}
}