package main

import (
	"context"
	"sync"
)

func concurrentExecute[T any](f func(T, context.Context, context.CancelFunc, chan error), items []T, concurrency int) (bool, error) {
	var wg sync.WaitGroup
	sem := make(chan struct{}, concurrency) // limit concurrent executions
	wg.Add(len(items))
	// Use context to handle cancellation on error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	errCh := make(chan error, 1)
	for _, item := range items {
		go func(param T){
			sem <- struct{}{} // acquire a slot
			defer func() { <-sem }() // release the slot
			defer wg.Done()
			select {
				case <-ctx.Done():
					return
				default:
			}
			f(param, ctx, cancel, errCh)
		}(item)
	}
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
		case err := <-errCh:
			return false, err
		case <-done:
			return true, nil
	}
}