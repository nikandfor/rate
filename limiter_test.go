package rate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimiter(t *testing.T) {
	now := time.Now()
	l := NewLimiter(now, 2, 5)

	assert.True(t, l.Have(now, 4))
	assert.True(t, l.Take(now, 4))
	assert.Equal(t, float64(1), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	now = now.Add(time.Second)
	assert.True(t, l.Have(now, 3))
	assert.True(t, l.Take(now, 3))
	assert.Equal(t, float64(0), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	now = now.Add(time.Second)
	assert.False(t, l.Have(now, 3))
	assert.False(t, l.Take(now, 3))
	assert.Equal(t, float64(2), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	backInTime := now.Add(-time.Second / 2)
	assert.Equal(t, float64(2), l.Value(backInTime)) // should not go back in time

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	now = now.Add(4 * time.Second)
	assert.Equal(t, float64(5), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	assert.Equal(t, float64(2), l.Rate())
	assert.Equal(t, float64(5), l.Capacity())

	l.Update(now, 3, 6)

	assert.Equal(t, float64(3), l.Rate())
	assert.Equal(t, float64(6), l.Capacity())
	assert.Equal(t, float64(5), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	assert.Equal(t, time.Duration(0), l.Borrow(now, 3))
	assert.Equal(t, float64(2), l.Value(now))
	assert.Equal(t, 2*time.Second, l.Borrow(now, 8))
	assert.Equal(t, float64(-6), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	assert.Equal(t, float64(-6), l.Set(now, 1))
	assert.Equal(t, float64(1), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}

	assert.Equal(t, time.Second, l.Borrow(now, 4))
	l.Return(now, 4)
	assert.Equal(t, float64(1), l.Value(now))

	l.Return(now, 8)
	assert.Equal(t, l.Capacity(), l.Value(now))

	if t.Failed() {
		t.Logf("limiter: %+v", l)
		return
	}
}

func TestLimiterBorrow(t *testing.T) {
	now := time.Now()
	l := NewLimiter(now, 2, 6)

	assert.Equal(t, 0*time.Second, l.Borrow(now, 4))
	assert.Equal(t, float64(2), l.Value(now))

	assert.Equal(t, 1*time.Second, l.Borrow(now, 4))
	assert.Equal(t, float64(-2), l.Value(now))

	assert.Equal(t, 3*time.Second, l.Borrow(now, 4))
	assert.Equal(t, float64(-6), l.Value(now))
}
