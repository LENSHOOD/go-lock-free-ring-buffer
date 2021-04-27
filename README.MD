## Lock free ring buffer
This repo is an implementation of lock-free ring buffer built on Golang.

You can find more information about the implementation at my [blog post](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/#more).

### API
Below is the API and how to use it:
```go
type RingBuffer interface {
    Offer(interface{}) (success bool)
    Poll() (value interface{}, success bool)
    SingleProducerOffer(valueSupplier func() (v interface{}, finish bool))
    SingleConsumerPoll(valueConsumer func(interface{}))
}
```
Generally, we can simply use `Offer()` and `Poll()` to use it like normal queue. Any types are acceptable, but for the memory efficiency concern, pointer is a better choice to handover by the ring buffer, no GC will happen during read/write.

We can create one as follows:
```go
import (
    "lfring"
)

buffer := lfring.New(lfring.Hybrid, 16)
```

The first argument of `New()` is the type of ring buffer, I currently provide two implementations, they both have same behavior, the only difference is a few inner implementations.

The second argument is `capacity`, which defines how big the ring buffer is (the calculate unit is the number of slots).

#### Method for special circumstance
There's two other methods you may noticed at API:
- `SingleProducerOffer()`
- `SingleConsumerPoll()`

This two methods is to cover two special circumstances as: multi-producer single-consumer, and single-producer multi-consumer.

When the usage scenario is mpsc or spmc, use such method may improve the handover performance since in those scenarios we can safely discard some heavy operations.

However, the performance test([in my blog post](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/#more)) show that `SingleConsumerPoll()` do improve the mpsc performance to about 1.5x, but `SingleProducerOffer()` become slower at spmc.

### Performance
1. Two type of lock-free ring buffer compare with go-channel at different capacities
![](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/5.png)
2. Two type of lock-free ring buffer compare with go-channel at different producer / consumer ratio
![](https://lenshood.github.io/2021/04/19/lock-free-ring-buffer/6.png)