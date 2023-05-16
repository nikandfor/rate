package rate

import (
	"time"
)

type (
	// Limiter uses classic token bucket algorithm.
	// Bucket is full when created.
	Limiter struct {
		rate, cap float64

		last time.Time
		val  float64
	}
)

// NewLimiter creates new token bucket limiter. Bucket is full when created.
// Use it to limit to at most rate per second tokens with at most cap burst.
//
// Each operaion takes current time and updates the state as it was
// filled up continuously.
// If time goes backwards it's ignored as it was already accounted.
// Thus you can't take unused tokens from the past if bucket is full now.
//
// Limiter is not safe to use in parallel.
func NewLimiter(now time.Time, rate, cap float64, opts ...Option) *Limiter {
	l := &Limiter{
		rate: rate,
		cap:  cap,
		val:  cap,
		last: now,
	}

	for _, o := range opts {
		o.apply(l, now)
	}

	return l
}

// Update advances Limiter state and sets new rate and cap for the future operations.
func (l *Limiter) Update(now time.Time, rate, cap float64) {
	l.advance(now)

	l.rate = rate
	l.cap = cap
}

// Have checks if Limiter have at least v tokens but don't take them.
func (l *Limiter) Have(now time.Time, v float64) bool {
	l.advance(now)

	return v <= l.val
}

// Take takes v tokens if there is enough of them.
// Take returns true if taken and false otherwise.
func (l *Limiter) Take(now time.Time, v float64) bool {
	l.advance(now)

	if v > l.val {
		return false
	}

	l.val -= v

	return true
}

// Borrow takes v tokens even if there is not enough of them and
// returns the time needed to wait before using them.
// If there was enough tokens the returned time is 0.
// Borrow can take even more tokens then burst capacity.
// Callers must check for Capacity on their own.
func (l *Limiter) Borrow(now time.Time, v float64) time.Duration {
	l.advance(now)

	l.val -= v

	if l.val >= 0 {
		return 0
	}

	return time.Duration(-l.val / l.rate * float64(time.Second))
}

// Return returns borrowed time back.
// If current value + returned v > Capacity it's truncated.
func (l *Limiter) Return(now time.Time, v float64) {
	l.val += v
	l.advance(now)
}

// Rate returns the current rate.
func (l *Limiter) Rate() float64 {
	return l.rate
}

// Capacity returns the current capacity.
func (l *Limiter) Capacity() float64 {
	return l.cap
}

// Value returns the current value.
// That is how much can we take at most in the moment now.
func (l *Limiter) Value(now time.Time) float64 {
	l.advance(now)

	return l.val
}

// Set sets current limiter state.
// It can be used to drain or fill the Limiter.
// Returns previous value.
func (l *Limiter) Set(now time.Time, v float64) float64 {
	x := l.val
	l.val = v
	l.last = now

	return x
}

// advance calculates accumulated value on time now.
// If value is more than capacity it's truncated.
// If time goes backwards, state is not changed.
func (l *Limiter) advance(now time.Time) {
	secs := now.Sub(l.last).Seconds()
	if secs < 0 {
		return
	}

	l.val += l.rate * secs
	l.last = now

	if l.val > l.cap {
		l.val = l.cap
	}
}
