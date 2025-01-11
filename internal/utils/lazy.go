package utils

import (
	"sync"
	"sync/atomic"
)

func NewLazy[T any](f func() (*T, error)) Lazy[T] {
	return Lazy[T]{f: f}
}

type Lazy[T any] struct {
	f func() (*T, error)

	ptr atomic.Pointer[T]
	mu  sync.Mutex
}

func (l *Lazy[T]) Get() (*T, error) {
	// `getSlow` is called so that the Get() function can be inlined
	// and potentially run faster after the initialization.
	// A similar approach is taken in the `sync.Once` Get() method
	if p := l.ptr.Load(); p == nil {
		return l.getSlow()
	} else {
		return p, nil
	}
}

func (l *Lazy[T]) getSlow() (*T, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.ptr.Load() == nil {
		v, err := l.f()
		if err == nil {
			l.ptr.Store(v)
		}
		return v, err
	}
	return l.ptr.Load(), nil
}
