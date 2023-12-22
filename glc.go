package glc

import (
	"context"
	"sync"
	"sync/atomic"
)

// WithContext executes the function `f` with the dynamic context bound to `ctx`.
// Calls to `GetContext` within `f` or the functions which `f` calls will return
// `ctx`.
//
// The dynamic binding does not cross goroutine boundaries, so these bindings are
// not visible to functions called with the `go` keyword.
func WithContext(ctx context.Context, f func()) {
	id := nextID()
	idmap.Store(id, ctx)
	defer idmap.Delete(id)
	encstart(id, f)
}

// GetContext returns the `context.Context` currently bound to the stack by
// `WithContext`.
func GetContext() context.Context {
	id, ok := lastID()
	if !ok {
		return nil
	}
	ctx, ok := idmap.Load(id)
	if !ok {
		return nil
	}
	return ctx
}

var id uint64
var idmap syncMap[uint64, context.Context]

func nextID() uint64 {
	return atomic.AddUint64(&id, 1)
}

type syncMap[T, U any] struct {
	m sync.Map
}

func (s *syncMap[T, U]) Store(key T, value U) {
	s.m.Store(key, value)
}

func (s *syncMap[T, U]) Load(key T) (U, bool) {
	var ret U
	v, ok := s.m.Load(key)
	if !ok {
		return ret, ok
	}
	ret = v.(U)
	return ret, ok
}

func (s *syncMap[T, U]) Delete(key T) {
	s.m.Delete(key)
}
