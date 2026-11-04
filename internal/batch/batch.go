// Package batch runs independent named tasks with bounded concurrency.
package batch

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
)

type Status string

const (
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

type Result struct {
	Name   string
	Status Status
	Err    error
}

// Run visits every name unless ctx is canceled. A callback failure does not
// cancel other callbacks. Errors in results belong to tasks; the returned error
// reports invalid Run arguments before any callback starts.
func Run(ctx context.Context, names []string, parallel int, execute func(context.Context, string) error) ([]Result, error) {
	if ctx == nil || execute == nil || parallel <= 0 {
		return nil, errors.New("batch requires a context, callback, and positive parallelism")
	}
	names = append([]string(nil), names...)
	sort.Strings(names)
	results := make([]Result, len(names))
	for i, name := range names {
		if strings.TrimSpace(name) == "" || (i > 0 && names[i-1] == name) {
			return nil, errors.New("batch task names must be nonempty and unique")
		}
		results[i] = Result{
			Name:   name,
			Status: StatusCanceled,
		}
	}
	var workers sync.WaitGroup
	jobs := make(chan int)
	for range min(parallel, len(names)) {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					results[index].Err = ctx.Err()
					continue
				}
				err := invoke(ctx, names[index], execute)
				results[index].Err = err
				switch {
				case err == nil:
					results[index].Status = StatusSucceeded
				case ctx.Err() != nil:
					results[index].Status = StatusCanceled
				default:
					results[index].Status = StatusFailed
				}
			}
		}()
	}
dispatch:
	for index := range names {
		if ctx.Err() != nil {
			break
		}
		select {
		case <-ctx.Done():
			break dispatch
		case jobs <- index:
		}
	}
	close(jobs)
	workers.Wait()
	if err := ctx.Err(); err != nil {
		for i := range results {
			if results[i].Status == StatusCanceled && results[i].Err == nil {
				results[i].Err = err
			}
		}
	}
	return results, nil
}

func invoke(ctx context.Context, name string, execute func(context.Context, string) error) (err error) {
	defer func() {
		if recover() != nil {
			err = errors.New("batch callback panicked")
		}
	}()
	return execute(ctx, name)
}
