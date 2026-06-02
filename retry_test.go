package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

var errFake = errors.New("fake error")

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	calls := 0
	err := Do(context.Background(), DefaultConfig(), func() error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndSucceeds(t *testing.T) {
	calls := 0
	cfg := Config{MaxAttempts: 3, Delay: time.Millisecond, MaxDelay: time.Second, Multiplier: 1}
	err := Do(context.Background(), cfg, func() error {
		calls++
		if calls < 3 {
			return errFake
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttempts(t *testing.T) {
	cfg := Config{MaxAttempts: 3, Delay: time.Millisecond, MaxDelay: time.Second, Multiplier: 1}
	err := Do(context.Background(), cfg, func() error { return errFake })
	if !errors.Is(err, errFake) {
		t.Fatalf("expected errFake, got %v", err)
	}
}

func TestDo_PermanentErrorStopsImmediately(t *testing.T) {
	calls := 0
	cfg := Config{MaxAttempts: 5, Delay: time.Millisecond, MaxDelay: time.Second, Multiplier: 1}
	err := Do(context.Background(), cfg, func() error {
		calls++
		return Permanent(errFake)
	})
	if !errors.Is(err, errFake) {
		t.Fatalf("expected errFake, got %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := Do(ctx, DefaultConfig(), func() error { return errFake })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestIsPermanent(t *testing.T) {
	if IsPermanent(errFake) {
		t.Fatal("plain error should not be permanent")
	}
	if !IsPermanent(Permanent(errFake)) {
		t.Fatal("wrapped error should be permanent")
	}
}
