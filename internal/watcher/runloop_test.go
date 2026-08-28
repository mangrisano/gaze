package watcher

import (
	"context"
	"testing"
	"time"
)

func TestRunLoopCancelsThePreviousRun(t *testing.T) {
	trigger := make(chan string)
	starts := make(chan struct{}, 10)
	cancels := make(chan struct{}, 10)

	// A fake run that blocks until its context is cancelled, recording when it
	// starts and when it is cancelled.
	run := func(ctx context.Context, _ string) {
		starts <- struct{}{}
		<-ctx.Done()
		cancels <- struct{}{}
	}

	go runLoop(trigger, run)

	for range 3 {
		trigger <- "x"
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

func TestRunLoopWaitsForThePreviousRun(t *testing.T) {
	trigger := make(chan string)
	events := make(chan string, 10)

	// Records "start" when a run begins and "end" when its context is cancelled.
	run := func(ctx context.Context, _ string) {
		events <- "start"
		<-ctx.Done()
		events <- "end"
	}

	go runLoop(trigger, run)

	trigger <- "x" // run 1
	trigger <- "x" // cancels run 1; must wait for its end before run 2 starts
	close(trigger) // cancels run 2

	// A run must fully finish before the next one starts (no overlap), so the
	// events strictly alternate.
	want := []string{"start", "end", "start", "end"}
	for i, w := range want {
		select {
		case got := <-events:
			if got != w {
				t.Fatalf("event %d = %q, want %q (runs must not overlap)", i, got, w)
			}
		case <-time.After(time.Second):
			t.Fatalf("timed out waiting for event %d (%q)", i, w)
		}
	}
}

func TestRunLoopPassesThePath(t *testing.T) {
	trigger := make(chan string)
	got := make(chan string, 1)

	run := func(ctx context.Context, path string) {
		got <- path
		<-ctx.Done()
	}

	go runLoop(trigger, run)
	trigger <- "internal/watcher/match.go"
	close(trigger)

	select {
	case p := <-got:
		if p != "internal/watcher/match.go" {
			t.Fatalf("run got path %q, want %q", p, "internal/watcher/match.go")
		}
	case <-time.After(time.Second):
		t.Fatal("run did not receive the path")
	}
}
