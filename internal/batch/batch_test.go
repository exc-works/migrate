package batch

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"testing"
	"time"
)

func TestBoundedOverlapAndCompleteTraversal(t *testing.T) {
	const parallel = 4
	names := make([]string, 1000)
	for i := range names {
		names[i] = fmt.Sprintf("target-%04d", len(names)-i)
	}
	original := append([]string(nil), names...)
	var active, peak, calls atomic.Int32
	started := make(chan struct{}, parallel)
	release := make(chan struct{})
	done := make(chan []Result, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	go func() {
		results, err := Run(ctx, names, parallel, func(context.Context, string) error {
			n := active.Add(1)
			defer active.Add(-1)
			for p := peak.Load(); n > p; p = peak.Load() {
				if peak.CompareAndSwap(p, n) {
					break
				}
			}
			if calls.Add(1) <= parallel {
				started <- struct{}{}
			}
			select {
			case <-release:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
		if err != nil {
			done <- nil
			return
		}
		done <- results
	}()
	for range parallel {
		select {
		case <-started:
		case <-ctx.Done():
			t.Fatal("callbacks did not overlap")
		}
	}
	close(release)
	select {
	case results := <-done:
		if len(results) != len(names) || calls.Load() != int32(len(names)) || peak.Load() != parallel || active.Load() != 0 {
			t.Fatalf("incomplete or unbounded traversal: results=%d calls=%d peak=%d active=%d", len(results), calls.Load(), peak.Load(), active.Load())
		}
		for i, result := range results {
			if result.Status != StatusSucceeded || result.Err != nil || (i > 0 && results[i-1].Name >= result.Name) {
				t.Fatalf("invalid result: %+v", result)
			}
		}
	case <-ctx.Done():
		t.Fatal("batch failed to complete")
	}
	if !reflect.DeepEqual(names, original) {
		t.Fatal("Run modified caller's names")
	}
}

func TestFailureAndPanicDoNotStopTraversal(t *testing.T) {
	exampleError := errors.New("example callback failure")
	var calls atomic.Int32
	results, err := Run(context.Background(), []string{"c", "b", "a"}, 1, func(_ context.Context, name string) error {
		calls.Add(1)
		switch name {
		case "a":
			return exampleError
		case "b":
			panic("private-example-value")
		default:
			return nil
		}
	})
	if err != nil || calls.Load() != 3 {
		t.Fatalf("traversal stopped: %v %d", err, calls.Load())
	}
	if results[0].Status != StatusFailed || !errors.Is(results[0].Err, exampleError) || results[1].Status != StatusFailed || results[1].Err.Error() != "batch callback panicked" || results[2].Status != StatusSucceeded {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestCancellationStopsDispatchAndWaitsForActiveCallbacks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 2)
	done := make(chan []Result, 1)
	var calls, cleaned atomic.Int32
	go func() {
		results, _ := Run(ctx, []string{"a", "b", "c", "d"}, 2, func(ctx context.Context, _ string) error {
			calls.Add(1)
			defer cleaned.Add(1)
			started <- struct{}{}
			<-ctx.Done()
			return ctx.Err()
		})
		done <- results
	}()
	for range 2 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("callbacks did not start")
		}
	}
	cancel()
	select {
	case results := <-done:
		if len(results) != 4 || calls.Load() != 2 || cleaned.Load() != 2 {
			t.Fatalf("incorrect cancellation: calls=%d cleaned=%d", calls.Load(), cleaned.Load())
		}
		for _, result := range results {
			if result.Status != StatusCanceled || !errors.Is(result.Err, context.Canceled) {
				t.Fatalf("invalid canceled result: %+v", result)
			}
		}
	case <-time.After(5 * time.Second):
		t.Fatal("batch did not wait for callback cleanup")
	}
}

func TestRunValidation(t *testing.T) {
	callback := func(context.Context, string) error { t.Fatal("invalid Run invoked callback"); return nil }
	for _, names := range [][]string{{"a", "a"}, {""}, {" "}} {
		if _, err := Run(context.Background(), names, 2, callback); err == nil {
			t.Fatal("accepted invalid names")
		}
	}
	for _, parallel := range []int{0, -1} {
		if _, err := Run(context.Background(), []string{"a"}, parallel, callback); err == nil {
			t.Fatal("accepted invalid parallelism")
		}
	}
	if _, err := Run(nil, nil, 1, callback); err == nil {
		t.Fatal("accepted nil context")
	}
	if _, err := Run(context.Background(), nil, 1, nil); err == nil {
		t.Fatal("accepted nil callback")
	}
	results, err := Run(context.Background(), nil, 1, callback)
	if err != nil || len(results) != 0 {
		t.Fatalf("empty traversal: %+v %v", results, err)
	}
}

func TestExpiredDeadlinePreservesReasonForUnstartedTargets(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()
	results, err := Run(ctx, []string{"a", "b", "c"}, 2, func(context.Context, string) error {
		t.Fatal("expired batch started a callback")
		return nil
	})
	if err != nil || len(results) != 3 {
		t.Fatalf("expired batch: %+v %v", results, err)
	}
	for _, result := range results {
		if result.Status != StatusCanceled || !errors.Is(result.Err, context.DeadlineExceeded) {
			t.Fatalf("lost deadline reason: %+v", result)
		}
	}
}

func TestCancellationPreservesCallbackCleanupErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cleanupError := errors.New("example cleanup failure")
	results, err := Run(ctx, []string{"a", "b"}, 1, func(context.Context, string) error {
		cancel()
		return errors.Join(ctx.Err(), cleanupError)
	})
	if err != nil || len(results) != 2 {
		t.Fatalf("canceled batch: %+v %v", results, err)
	}
	if !errors.Is(results[0].Err, context.Canceled) || !errors.Is(results[0].Err, cleanupError) || !errors.Is(results[1].Err, context.Canceled) {
		t.Fatal("cancellation replaced callback or unstarted errors")
	}
}

func TestMoreWorkersThanTargets(t *testing.T) {
	for _, names := range [][]string{{"shard_0"}, {"shard_0", "shard_1"}} {
		var calls atomic.Int32
		results, err := Run(context.Background(), names, 8, func(context.Context, string) error {
			calls.Add(1)
			return nil
		})
		if err != nil || len(results) != len(names) || int(calls.Load()) != len(names) {
			t.Fatalf("incorrect small batch: %+v calls=%d err=%v", results, calls.Load(), err)
		}
		for i, result := range results {
			if result.Name != names[i] || result.Status != StatusSucceeded || result.Err != nil {
				t.Fatalf("incorrect target result: %+v", result)
			}
		}
	}
}
