package main

import "time"

// debounce forwards one signal only after in has been quiet for d, coalescing a
// burst into a single output. It closes the returned channel when in closes.
func debounce(in <-chan struct{}, d time.Duration) <-chan struct{} {
	out := make(chan struct{})
	go func() {
		defer close(out)

		// nil = disarmed: a nil channel never fires in a select.
		var timer <-chan time.Time

		for {
			select {
			case _, ok := <-in:
				if !ok {
					return
				}
				timer = time.After(d)
			case <-timer:
				out <- struct{}{}
				timer = nil
			}
		}
	}()
	return out
}
