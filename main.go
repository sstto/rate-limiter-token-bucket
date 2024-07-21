package main

import (
	"fmt"
	"time"
)

func main() {
	bucket, _ := NewBucketBuilder().
		SetName("my bucket").
		SetCapacity(1000).
		SetRefillTokens(100).
		SetRefillPeriod(time.Second).
		Build()

	ok := true
	count := 0
	for ok {
		time.Sleep(5 * time.Millisecond)
		ok = bucket.TryConsume()
		count++
	}
	fmt.Println("phase 1 count: ", count)

	time.Sleep(5 * time.Second)

	count = 0
	ok = true
	for ok {
		time.Sleep(5 * time.Millisecond)
		ok = bucket.TryConsume()
		count++
	}
	fmt.Println("phase 2 count: ", count)

	bucket.Close()
	<-bucket.done
	time.Sleep(5 * time.Second)
}
