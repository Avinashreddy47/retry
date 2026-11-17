# retry

[![Go Reference](https://pkg.go.dev/badge/github.com/Avinashreddy47/retry.svg)](https://pkg.go.dev/github.com/Avinashreddy47/retry)
[![Go Report Card](https://goreportcard.com/badge/github.com/Avinashreddy47/retry)](https://goreportcard.com/report/github.com/Avinashreddy47/retry)
[![Go](https://github.com/Avinashreddy47/retry/actions/workflows/go.yml/badge.svg)](https://github.com/Avinashreddy47/retry/actions/workflows/go.yml)

A small, context-aware retry utility for Go with exponential backoff and
optional observability hooks.

## Features

- Context cancellation and deadlines
- Configurable exponential backoff with a maximum delay
- Permanent errors that stop retrying immediately
- Retry events for logs, metrics, and traces
- Zero external dependencies

## Install

```bash
go get github.com/Avinashreddy47/retry
```

## Usage

```go
import "github.com/Avinashreddy47/retry"

// Use defaults: 3 attempts, 100ms → 200ms → capped at 10s
err := retry.Do(ctx, retry.DefaultConfig(), func() error {
    return callUnreliableAPI()
})

// Custom config
err := retry.Do(ctx, retry.Config{
    MaxAttempts: 5,
    Delay:       50 * time.Millisecond,
    MaxDelay:    2 * time.Second,
    Multiplier:  2.0,
}, func() error {
    return callUnreliableAPI()
})

// Stop retrying immediately on certain errors
err := retry.Do(ctx, retry.DefaultConfig(), func() error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    if resp.StatusCode == 400 {
        return retry.Permanent(fmt.Errorf("bad request: will not retry"))
    }
    return nil
})
```

## Observability

Use `OnRetry` to record a failed attempt before the next attempt begins:

```go
cfg := retry.DefaultConfig()
cfg.OnRetry = func(event retry.Event) {
    slog.Warn("operation failed; retrying",
        "attempt", event.Attempt,
        "error", event.Err,
        "delay", event.Delay,
    )
}

err := retry.Do(ctx, cfg, func() error {
    return callUnreliableAPI()
})
```

The callback is not invoked after a successful attempt, a permanent error, or
the final failed attempt because no retry will follow.

## Behavior

| Situation | Result |
|---|---|
| `fn` returns `nil` | stops, returns `nil` |
| `fn` returns `retry.Permanent(err)` | stops immediately, returns unwrapped `err` |
| `ctx` cancelled / timed out | stops, returns `ctx.Err()` |
| All attempts exhausted | returns last error from `fn` |

## Compatibility

The module requires Go 1.22 or newer and has no third-party dependencies.

## License

MIT
