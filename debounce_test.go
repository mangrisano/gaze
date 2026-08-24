package main

import (
	"testing"
	"time"
)

func TestDebounceCoalescesBursts(t *testing.T) {
	in := make(chan struct{})
	out := debounce(in, 40*time.Millisecond)

	go func() {
		// First burst: 5 events well within the 40ms window -> one run.
		for range 5 {
			in <- struct{}{}
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(80 * time.Millisecond) // let the window elapse

		// Second, separate burst -> another run.
		in <- struct{}{}
		time.Sleep(80 * time.Millisecond)

		close(in)
	}()

	runs := 0
	for range out {
		runs++
	}
	if runs != 2 {
		t.Fatalf("got %d runs, want 2", runs)
	}
}
