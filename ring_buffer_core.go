package go_mpsc_ring_buffer

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

type ringBufferBasement struct {
	head uint64
	tail uint64
	capacity uint64
	mask uint64
}