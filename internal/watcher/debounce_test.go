package watcher

import (
	"slices"
	"testing"
	"time"
)

func TestDebounceCoalescesBursts(t *testing.T) {
	in := make(chan string)
	out := debounce(in, 40*time.Millisecond)

	go func() {
		// First burst: 5 events well within the 40ms window -> one run,
		// carrying the last path of the burst.
		for _, p := range []string{"a", "b", "c", "d", "e"} {
			in <- p
			time.Sleep(5 * time.Millisecond)
		}
		time.Sleep(80 * time.Millisecond) // let the window elapse

		// Second, separate burst -> another run.
		in <- "f"
		time.Sleep(80 * time.Millisecond)

		close(in)
	}()

	var got []string
	for p := range out {
		got = append(got, p)
	}
	want := []string{"e", "f"} // one run per burst, last path wins
	if !slices.Equal(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
