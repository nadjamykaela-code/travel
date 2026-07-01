package service

import (
	"sync/atomic"
)

type Limiter struct {
	remaining atomic.Int64
}

func NewLimiter(max int) *Limiter {
	l := &Limiter{}
	l.remaining.Store(int64(max))
	return l
}

func (l *Limiter) Allow() bool {
	rem := l.Remaining()
	if rem <= 0 {
		return false
	}
	l.remaining.Add(-1)
	return true
}

func (l *Limiter) Remaining() int {
	return int(l.remaining.Load())
}
