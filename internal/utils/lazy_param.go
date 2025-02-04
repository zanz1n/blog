package utils

import (
	"sync"
	"sync/atomic"
)

func NewLazyParam[T any, P any](f func(P) (*T, error)) LazyParam[T, P] {
	return LazyParam[T, P]{f: f}
}

type LazyParam[T any, P any] struct {
	f func(P) (*T, error)

	ptr atomic.Pointer[T]
	mu  sync.Mutex
}

func (l *LazyParam[T, P]) Get(param P) (*T, error) {
	// `getSlow` is called so that the Get() function can be inlined
	// and potentially run faster after the initialization.
	// A similar approach is taken in the `sync.Once` Get() method
	if p := l.ptr.Load(); p == nil {
		return l.getSlow(param)
	} else {
		return p, nil
	}
}

func (l *LazyParam[T, P]) getSlow(param P) (*T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ptr.Load() == nil {
		v, err := l.f(param)
		if err == nil {
			l.ptr.Store(v)
		}
		return v, err
	}
	return l.ptr.Load(), nil
}
