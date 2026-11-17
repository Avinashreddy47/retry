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

func TestDo_OnRetryReceivesFailedAttempts(t *testing.T) {
	var events []Event
	calls := 0
	cfg := Config{
		MaxAttempts: 3,
		Delay:       time.Millisecond,
		MaxDelay:    time.Second,
		Multiplier:  2,
		OnRetry: func(event Event) {
			events = append(events, event)
		},
	}

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

	if len(events) != 2 {
		t.Fatalf("expected 2 retry events, got %d", len(events))
	}
	for i, event := range events {
		wantAttempt := i + 1
		wantDelay := time.Millisecond * time.Duration(1<<i)
		if event.Attempt != wantAttempt {
			t.Errorf("event %d: expected attempt %d, got %d", i, wantAttempt, event.Attempt)
		}
		if !errors.Is(event.Err, errFake) {
			t.Errorf("event %d: expected errFake, got %v", i, event.Err)
		}
		if event.Delay != wantDelay {
			t.Errorf("event %d: expected delay %v, got %v", i, wantDelay, event.Delay)
		}
	}
}

func TestDo_OnRetryNotCalledWithoutAnotherAttempt(t *testing.T) {
	tests := []struct {
		name string
		fn   func() error
	}{
		{name: "success", fn: func() error { return nil }},
		{name: "permanent error", fn: func() error { return Permanent(errFake) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			cfg := DefaultConfig()
			cfg.OnRetry = func(Event) { calls++ }

			_ = Do(context.Background(), cfg, tt.fn)
			if calls != 0 {
				t.Fatalf("expected no retry events, got %d", calls)
			}
		})
	}
}

func TestDo_ContextCancellationDuringBackoff(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := Config{
		MaxAttempts: 3,
		Delay:       time.Hour,
		MaxDelay:    time.Hour,
		Multiplier:  1,
		OnRetry:     func(Event) { cancel() },
	}

	err := Do(ctx, cfg, func() error { return errFake })
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestNextDelay(t *testing.T) {
	tests := []struct {
		name       string
		attempt    int
		multiplier float64
		max        time.Duration
		want       time.Duration
	}{
		{name: "initial delay", attempt: 0, multiplier: 2, max: time.Second, want: 100 * time.Millisecond},
		{name: "exponential delay", attempt: 2, multiplier: 2, max: time.Second, want: 400 * time.Millisecond},
		{name: "capped delay", attempt: 4, multiplier: 2, max: 500 * time.Millisecond, want: 500 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextDelay(100*time.Millisecond, tt.multiplier, tt.max, tt.attempt)
			if got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
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
