package watcher

import "time"

// debounce forwards one signal only after in has been quiet for d, coalescing a
// burst into a single output. It closes the returned channel when in closes.
func debounce(in <-chan string, d time.Duration) <-chan string {
	out := make(chan string)
	var last string
	go func() {
		defer close(out)

		// nil = disarmed: a nil channel never fires in a select.
		var timer <-chan time.Time

		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				last = v
				timer = time.After(d)
			case <-timer:
				out <- last
				timer = nil
			}
		}
	}()
	return out
}
