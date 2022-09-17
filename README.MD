## Lock free ring buffer  
[![Go Report Card](https://goreportcard.com/badge/github.com/LENSHOOD/go-lock-free-ring-buffer)](https://goreportcard.com/report/github.com/LENSHOOD/go-lock-free-ring-buffer) [![Go Reference](https://pkg.go.dev/badge/github.com/LENSHOOD/go-lock-free-ring-buffer.svg)](https://pkg.go.dev/github.com/LENSHOOD/go-lock-free-ring-buffer) [![GitHub license](https://img.shields.io/github/license/LENSHOOD/go-lock-free-ring-buffer)](https://github.com/LENSHOOD/go-lock-free-ring-buffer/blob/master/LICENSE) [![GitHub release](https://img.shields.io/github/tag/LENSHOOD/go-lock-free-ring-buffer)](https://github.com/LENSHOOD/go-lock-free-ring-buffer/releases/tag/v0.1.0)

This repo is an implementation of lock-free ring buffer built on Golang.

You can find more information about this implementation at my [blog post](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/#more).

Please note that v0.2.0 involved [generics feature](https://go.dev/doc/go1.18#generics), which requires go version >= 1.18. Lower go version please choose v0.1.0.

### API
Below is the API and how to use it:
```go
type RingBuffer[T any] interface {
  Offer(T) (success bool)
  Poll() (value T, success bool)
}
```
We can simply call `Offer()` and `Poll()` to use it like a normal queue. 

The GCShape introduced by generics feature can ensure that no heap memory allocation during `Offer()` and `Poll()`. [Here](https://lenshood.github.io/2022/08/01/optimize-lfring-performance/) is an article to explain the performance changes before and after involve generic.

When create an instance, say we want to use it to store `string`:
```go
import (
  "github.com/LENSHOOD/go-lock-free-ring-buffer"
)

buffer := lfring.New[string](lfring.NodeBased, 16)
```

The first argument of `New()` is the type of ring buffer, I currently provide two implementations, they both have same behavior, but benchmark test shows that the "NodeBased" one has better performance.

The second argument `capacity` defines how big the ring buffer is, in consideration of different concrete type, the size of buffer maybe different. For instance, string has two underlying elements `str unsafe.Pointer` and `len int`, so if we build a buffer has `capacity=16`, the size of buffer array will be `16*(8+8)=256 bytes`(64bit platform).

### Performance
1. Two types of lock-free ring buffer compare with go channel in different capacities
![](https://github.com/LENSHOOD/lenshood.github.io/blob/source/source/_posts/decide-lfring-channel/capacity-all.png?raw=true)

2. Two types of lock-free ring buffer compare with go channel in different producers / consumers ratio
![](https://github.com/LENSHOOD/lenshood.github.io/blob/source/source/_posts/decide-lfring-channel/thread-all.png?raw=true)

3. Two types of lock-free ring buffer compare with go channel in different threads
![](https://github.com/LENSHOOD/lenshood.github.io/blob/source/source/_posts/decide-lfring-channel/producer-all.png?raw=true)

From above performance curve, we can see that ring buffer get better performance under some specific conditions. Below image clarifies the proper use scenario of ring buffer.
![](https://github.com/LENSHOOD/lenshood.github.io/blob/source/source/_posts/decide-lfring-channel/3d-capacity-producer.png?raw=true)

As the result, ring buffer fits better in the case of producers/consumers not balance, and capacity not too big.

Above images are screenshots, check full charts [here](https://lenshood.github.io/2022/09/04/decide-lfring-channel/).

### Unfinished features
- [ ] Try to optimize the performance of single producer/consumer performance

    - Based on the previous result, the ratio of production/consumption speed plays a partial key role to influence ordinary channel's performance. We can move one step forward, to do some optimization at such a special case of single producer/consumer. 

    - Single means no contention, so the costly CAS operation may be replaced with normal operation, and single load/store may be replaced with vectorized load/store. Furthermore, once we say vectorization, we think introduce SIMD to accelerate load/store operations.