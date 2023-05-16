package rate

import "time"

type (
	Option interface {
		apply(l *Limiter, now time.Time)
	}

	OptionFunc func(l *Limiter, now time.Time)
)

// WithValue is an Option for NewLimiter that sets initial value.
func WithValue(v float64) Option {
	return OptionFunc(func(l *Limiter, now time.Time) {
		l.Set(now, v)
	})
}

func (f OptionFunc) apply(l *Limiter, now time.Time) {
	f(l, now)
}
