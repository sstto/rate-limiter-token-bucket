package bucket

import (
	"fmt"
	"time"
)

// bucket builder
type Builder struct {
	name         string
	capacity     int
	refillTokens int
	refillPeriod time.Duration
	ch           chan struct{}
}

func NewBuilder() *Builder {
	return &Builder{ // default value
		name:         "default",
		capacity:     100,
		refillTokens: 10,
		refillPeriod: time.Second,
	}
}

func (b *Builder) SetName(n string) *Builder {
	b.name = n
	return b
}

func (b *Builder) SetCapacity(c int) *Builder {
	b.capacity = c
	return b
}

func (b *Builder) SetRefillTokens(r int) *Builder {
	b.refillTokens = r
	return b
}

func (b *Builder) SetRefillPeriod(p time.Duration) *Builder {
	b.refillPeriod = p
	return b
}

func (b *Builder) Build() (*bucket, error) {
	if b.capacity <= 0 {
		return nil, fmt.Errorf("invalid capacity: %d", b.capacity)
	}
	// TODO: validation 추가.

	ch := make(chan struct{}, b.capacity)
	done := make(chan struct{})
	// fill a channel to its capacity
	for i := 0; i < b.capacity; i++ {
		ch <- struct{}{}
	}
	fmt.Println("fill a channel to its capacity. name: ", b.name)

	go func(ch chan struct{}, n string, r int, p time.Duration) {
		defer func() {
			close(ch)
			fmt.Println("token refill goroutine has terminated. name: ", n)
		}()

		for {
			select {
			case <-done:
				return
			default:
				time.Sleep(p)
				refillTokens(ch, r)
			}
		}
	}(ch, b.name, b.refillTokens, b.refillPeriod)

	return &bucket{
		name:         b.name,
		capacity:     b.capacity,
		refillTokens: b.refillTokens,
		refillPeriod: b.refillPeriod,
		ch:           ch,
		Done:         done,
	}, nil
}

// bucket
type bucket struct {
	name         string
	capacity     int
	refillTokens int
	refillPeriod time.Duration
	ch           chan struct{}
	Done         chan struct{}
}

func (b *bucket) TryConsume() bool {
	select {
	case <-b.ch:
		return true
	default:
		return false
	}
}

func (b *bucket) Close() {
	fmt.Println("close bucket. name: ", b.name)
	close(b.Done)
}

func refillTokens(ch chan struct{}, r int) {
	for i := 0; i < r; i++ {
		refillOrDrop(ch)
	}
	fmt.Printf("refill %v token\n", r)
}

func refillOrDrop(ch chan struct{}) {
	select {
	case ch <- struct{}{}: // value sent succesfully
	default: // value dropped
	}
}
