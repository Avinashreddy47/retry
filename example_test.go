package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Avinashreddy47/retry"
)

func Example() {
	calls := 0
	err := retry.Do(context.Background(), retry.DefaultConfig(), func() error {
		calls++
		if calls < 3 {
			return errors.New("not ready yet")
		}
		return nil
	})
	fmt.Println(err)
	// Output: <nil>
}

func ExamplePermanent() {
	err := retry.Do(context.Background(), retry.DefaultConfig(), func() error {
		return retry.Permanent(errors.New("bad request"))
	})
	fmt.Println(err)
	// Output: bad request
}

func ExampleConfig() {
	cfg := retry.Config{
		MaxAttempts: 5,
		Delay:       50 * time.Millisecond,
		MaxDelay:    2 * time.Second,
		Multiplier:  2.0,
	}

	err := retry.Do(context.Background(), cfg, func() error {
		return nil
	})
	fmt.Println(err)
	// Output: <nil>
}
