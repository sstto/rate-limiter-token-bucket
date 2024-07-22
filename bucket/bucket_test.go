package bucket

import (
	"testing"
	"time"
)

func TestBuilder(t *testing.T) {
	b, err := NewBuilder().
		SetName("test-builder").
		SetCapacity(5).
		SetRefillTokens(1).
		SetRefillPeriod(500 * time.Millisecond).
		Build()
	if err != nil {
		t.Fatalf("버킷 생성 실패: %v", err)
	}
	defer b.Close()

	if b == nil {
		t.Fatalf("버킷이 nil입니다.")
	}
	if b.capacity != 5 {
		t.Errorf("기대된 용량: 5, 실제 용량: %d", b.capacity)
	}
	if b.refillTokens != 1 {
		t.Errorf("기대된 리필 토큰: 1, 실제 리필 토큰: %d", b.refillTokens)
	}
	if b.refillPeriod != 500*time.Millisecond {
		t.Errorf("기대된 리필 주기: %v, 실제 리필 주기: %v", 500*time.Millisecond, b.refillPeriod)
	}
}

func TestBucket(t *testing.T) {
	b, err := NewBuilder().
		SetName("test").
		SetCapacity(5).
		SetRefillTokens(1).
		SetRefillPeriod(500 * time.Millisecond).
		Build()
	if err != nil {
		t.Fatalf("버킷 생성 실패: %v", err)
	}
	defer b.Close()

	// 초기 토큰 소비
	for i := 0; i < 5; i++ {
		if !b.TryConsume() {
			t.Fatalf("토큰 소비 실패: %d", i)
		}
	}

	// 토큰이 부족할 경우
	if b.TryConsume() {
		t.Fatalf("토큰이 부족해야 함에도 소비됨")
	}

	// 리필 대기
	time.Sleep(600 * time.Millisecond)

	// 리필 후 토큰 소비
	if !b.TryConsume() {
		t.Fatalf("리필 후 토큰 소비 실패")
	}
}

func TestBucketCapacity(t *testing.T) {
	b, err := NewBuilder().
		SetName("test-capacity").
		SetCapacity(3).
		SetRefillTokens(1).
		SetRefillPeriod(time.Second).
		Build()
	if err != nil {
		t.Fatalf("버킷 생성 실패: %v", err)
	}
	defer b.Close()

	// 초기 토큰 소비
	for i := 0; i < 3; i++ {
		if !b.TryConsume() {
			t.Fatalf("토큰 소비 실패: %d", i)
		}
	}

	// 채널이 가득 찼을 때 추가 리필 확인
	for i := 0; i < 3; i++ {
		b.TryConsume()
	}

	time.Sleep(1200 * time.Millisecond)

	if !b.TryConsume() {
		t.Fatalf("리필 후 토큰 소비 실패")
	}
}

func TestRefillTokens(t *testing.T) {
	ch := make(chan struct{}, 5)

	// 채널을 초기 토큰으로 채우기
	for i := 0; i < 3; i++ {
		ch <- struct{}{}
	}

	// 리필 테스트
	refillTokens(ch, 2)

	// 채널에 리필된 토큰 확인
	if len(ch) != 5 {
		t.Fatalf("기대된 토큰 개수: 5, 실제 토큰 개수: %d", len(ch))
	}

	// 채널이 가득 찼을 때 추가 리필 확인
	refillTokens(ch, 2)

	// 채널에 토큰이 넘치지 않는지 확인
	if len(ch) != 5 {
		t.Fatalf("기대된 토큰 개수: 5, 실제 토큰 개수: %d", len(ch))
	}
}
