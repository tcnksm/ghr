package retry

import "time"

// Retry calls the `fn` and if it returns the error, retry to call `fn` after `interval` duration.
// The `fn` is called up to `n` times.
func Retry(n uint, interval time.Duration, fn func() error) (err error) {
	for n > 0 {
		n--
		err = fn()
		if err == nil || n <= 0 {
			break
		}
		time.Sleep(interval)
	}
	return err
}
