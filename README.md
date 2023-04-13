[![Documentation](https://pkg.go.dev/badge/github.com/nikandfor/rate)](https://pkg.go.dev/github.com/nikandfor/rate?tab=doc)
[![CircleCI](https://circleci.com/gh/nikandfor/rate.svg?style=svg)](https://circleci.com/gh/nikandfor/rate)
[![codecov](https://codecov.io/gh/nikandfor/rate/branch/master/graph/badge.svg)](https://codecov.io/gh/nikandfor/rate)
[![GolangCI](https://golangci.com/badges/github.com/nikandfor/rate.svg)](https://golangci.com/r/github.com/nikandfor/rate)
[![Go Report Card](https://goreportcard.com/badge/github.com/nikandfor/rate)](https://goreportcard.com/report/github.com/nikandfor/rate)
![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/nikandfor/rate?sort=semver)

# Rate limiter

Token bucket rate limiter.

## Usage

Create limiter
```go
l := rate.NewLimiter(time.Now(),
	1000 / time.Second.Seconds(), // 1000 tokens per second
	2000, // 2000 tokens burst
	)

// smooth 1KB per second with at most 128 bytes at a time
l = rate.NewLimiter(time.Now(),
	1000 / time.Second.Seconds(),
	128)

// 3 MB per minute allowing to spend it all at once
l = rate.NewLimiter(time.Now(),
	3000000 / time.Minute.Seconds(),
	3000000)
```

Take or drop
```go
func (c *Conn) Write(p []byte) (int, error) {
	if !l.Take(time.Now(), len(p)) {
		return 0, ErrLimited
	}

	return c.Conn.Write(p)
}
```

Borrow and wait
```go
func (c *Conn) Write(p []byte) (int, error) {
	delay := l.Borrow(time.Now(), len(p))

	if delay != 0 {
		time.Sleep(delay)
	}

	return c.Conn.Write(p)
}
```

Write as much as we can
```go
func (c *Conn) Write(p []byte) (int, error) {
	now := time.Now()

	val := l.Value(now)

	n := int(val)
	if n > len(p) {
		n = len(p)
	}

	_ = l.Take(now, float64(n)) // must be true

	n, err := c.Conn.Write(p[:n])
	if err != nil {
		return n, err
	}
	if n != len(p) {
		err = ErrLimited
	}

	return n, err
}
```
