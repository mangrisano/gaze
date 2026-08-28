package watcher

import "context"

// runLoop runs run once per signal on trigger, cancelling the previous run when
// a new signal arrives so only one runs at a time. It returns when trigger closes.
func runLoop(trigger <-chan string, run func(ctx context.Context, path string)) {
	var cancel context.CancelFunc
	var done chan struct{}

	stop := func() {
		if cancel != nil {
			cancel()
			<-done
		}
	}
	for path := range trigger {
		stop()
		newCtx, cancFunc := context.WithCancel((context.Background()))
		cancel = cancFunc
		done = make(chan struct{})
		go func() {
			run(newCtx, path)
			close(done)
		}()
	}
	stop()
}
