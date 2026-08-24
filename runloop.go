package main

import "context"

// runLoop runs run once per signal on trigger, cancelling the previous run when
// a new signal arrives so only one runs at a time. It returns when trigger closes.
func runLoop(trigger <-chan struct{}, run func(ctx context.Context)) {
	var cancel context.CancelFunc

	for range trigger {
		if cancel != nil {
			cancel()
		}
		newCtx, cancFunc := context.WithCancel((context.Background()))
		cancel = cancFunc
		go run(newCtx)
	}
	if cancel != nil {
		cancel()
	}
}
