package middleware

import (
	"log"
	"strconv"
	"testing"
	"time"
)

func TestSlidingWindowCounter(t *testing.T) {
	sw, err := NewSlidingWindowCounter(
		3*time.Second,
		1*time.Second,
		30,
		nil,
		log.Default(),
	)
	if err != nil {
		t.Fatal(err)
	}

	key := "key"

	// distributed low request: 1 req per 500 millisecond => should success
	for range 12 {
		accepted := sw.Take(key)
		if !accepted {
			t.Error("should be accepted")
		}
		time.Sleep(500 * time.Millisecond)
	}

	time.Sleep(6 * time.Second)

	// burst for short period => should fail in 31th request
	for i := range 31 {
		accepted := sw.Take(key)
		if accepted != (i != 30) {
			t.Errorf("should not accept on 30, else accept: %d", i)
		}
	}

	time.Sleep(6 * time.Second)

	// distributed high request: 1 req per 70 millisecond => should fail in 31th request
	for i := range 31 {
		accepted := sw.Take(key)
		if accepted != (i != 30) {
			t.Errorf("should not accept on 30, else accept: %d", i)
		}
		time.Sleep(70 * time.Millisecond)
	}

	// other keys should not interfere each other => should success
	time.Sleep(6 * time.Second)
	for i := range 40 {
		key := strconv.Itoa(i % 2)
		accepted := sw.Take(key)
		if !accepted {
			t.Errorf("should accepted")
		}
	}
}
