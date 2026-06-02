# retry

[![Go Reference](https://pkg.go.dev/badge/github.com/Avinashreddy47/retry.svg)](https://pkg.go.dev/github.com/Avinashreddy47/retry)
[![Go Report Card](https://goreportcard.com/badge/github.com/Avinashreddy47/retry)](https://goreportcard.com/report/github.com/Avinashreddy47/retry)

A minimal, context-aware retry utility for Go with exponential backoff.

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

## Behavior

| Situation | Result |
|---|---|
| `fn` returns `nil` | stops, returns `nil` |
| `fn` returns `retry.Permanent(err)` | stops immediately, returns unwrapped `err` |
| `ctx` cancelled / timed out | stops, returns `ctx.Err()` |
| All attempts exhausted | returns last error from `fn` |

## License

MIT
