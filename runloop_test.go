package main

import (
	"context"
	"testing"
	"time"
)

func TestRunLoopCancelsThePreviousRun(t *testing.T) {
	trigger := make(chan struct{})
	starts := make(chan struct{}, 10)
	cancels := make(chan struct{}, 10)

	// A fake run that blocks until its context is cancelled, recording when it
	// starts and when it is cancelled.
	run := func(ctx context.Context) {
		starts <- struct{}{}
		<-ctx.Done()
		cancels <- struct{}{}
	}

	go runLoop(trigger, run)

	for range 3 {
		trigger <- struct{}{}
	}
	close(trigger)

	// All three runs must start...
	for range 3 {
		select {
		case <-starts:
		case <-time.After(time.Second):
			t.Fatal("a run did not start")
		}
	}
	// ...and each must be cancelled: runs 1 and 2 by the next trigger, run 3 by
	// the shutdown when trigger closes.
	for range 3 {
		select {
		case <-cancels:
		case <-time.After(time.Second):
			t.Fatal("a run was not cancelled")
		}
	}
}
