package rate_test

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/nikandfor/rate"
)

type WriterFunc func([]byte) (int, error)

func (f WriterFunc) Write(p []byte) (int, error) {
	return f(p)
}

func ExampleLimiter_Take_full() {
	// Mind we want to limit packets bandwidth.

	// Out test fake time
	base := time.Now()
	now := base

	// downstream writer
	w := io.Discard

	var ErrSpeedLimited = errors.New("speed limited")
	var r *rate.Limiter // will init later

	// the actual usage example of rate.Limiter
	limitedWrite := func(p []byte) (n int, err error) {
		if !r.Take(now, float64(len(p))) {
			return 0, ErrSpeedLimited
		}

		return w.Write(p)
	}

	fmt.Printf("Max burst of 2KiB with rate of 1KiB per second\n")

	tRate := float64(1*1024) / time.Second.Seconds() // 1KiB per second
	burst := float64(2 * 1024)                       // 2KiB at once at most

	r = rate.NewLimiter(now, tRate, burst)

	for i := 0; i < 4; i++ {
		n, err := limitedWrite(make([]byte, 1024))
		fmt.Printf("time %5v: %d %v\n", now.Sub(base), n, err)
		now = now.Add(time.Second / 2)
	}

	// Output:
	// Max burst of 2KiB with rate of 1KiB per second
	// time    0s: 1024 <nil>
	// time 500ms: 1024 <nil>
	// time    1s: 1024 <nil>
	// time  1.5s: 0 speed limited
}

func ExampleLimiter_Borrow_full() {
	// Mind we want to limit output speed for tcp connection.

	// Out test fake time
	base := time.Now()
	now := base

	// downstream writer
	w := WriterFunc(func(p []byte) (int, error) {
		fmt.Printf("%5d bytes written at %v\n", len(p), now.Sub(base))

		return len(p), nil
	})

	var r *rate.Limiter // will init later

	// the actual usage example of rate.Limiter
	limitedWrite := func(p []byte) (n int, err error) {
		var m int
		lim := len(p)

		if c := r.Capacity(); float64(lim) > c {
			lim = int(c)
		}

		for err == nil && n < len(p) {
			if lim > len(p)-n {
				lim = len(p) - n
			}

			delay := r.Borrow(now, float64(lim))
			if delay > 0 {
				now = now.Add(delay) // time.Sleep(delay)
			}

			m, err = w.Write(p[n : n+lim])
			n += m
		}

		return
	}

	fmt.Printf("Max burst of 515 bytes with rate of 1KiB per second\n")

	tRate := float64(1*1024) / time.Second.Seconds() // 1KiB per second
	burst := float64(512)                            // 128 bytes at once at most

	r = rate.NewLimiter(now, tRate, burst)

	_, _ = limitedWrite(make([]byte, 5*1024))

	// Output:
	// Max burst of 515 bytes with rate of 1KiB per second
	//   512 bytes written at 0s
	//   512 bytes written at 500ms
	//   512 bytes written at 1s
	//   512 bytes written at 1.5s
	//   512 bytes written at 2s
	//   512 bytes written at 2.5s
	//   512 bytes written at 3s
	//   512 bytes written at 3.5s
	//   512 bytes written at 4s
	//   512 bytes written at 4.5s
}
