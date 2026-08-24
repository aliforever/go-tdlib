package tdlib

import "context"

// Future is a one-shot asynchronous result, returned by DoAsync and Map.
// The result is buffered and memoized: complete happens exactly once and
// every Await observes the same value.
type Future[T any] struct {
	done chan struct{} // closed by complete; never nil on root futures
	v    T
	err  error

	// chain, when non-nil, makes this a derived future (see Map): Await
	// delegates to it instead of waiting on done. done is nil then.
	chain func(ctx context.Context) (T, error)
}

func newFuture[T any]() *Future[T] {
	return &Future[T]{done: make(chan struct{})}
}

// complete delivers the one-shot result. It is called exactly once per root
// future, by the goroutine that produces the value.
func (f *Future[T]) complete(v T, err error) {
	f.v, f.err = v, err
	close(f.done)
}

// Await blocks until the result is ready or ctx ends. Awaiting a derived
// future (from Map) awaits the parent and then applies the mapping
// function; the mapping runs once per Await call.
func (f *Future[T]) Await(ctx context.Context) (T, error) {
	if f.chain != nil {
		return f.chain(ctx)
	}
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-f.done:
		return f.v, f.err
	}
}

// Map derives a future whose result is fn applied to f's result. The
// composition is lazy: no goroutine is spawned and fn does not run until
// the derived future is awaited. If f fails, fn never runs and the error
// propagates; a fn error becomes the derived future's error.
func (f *Future[T]) Map[U any](fn func(T) (U, error)) *Future[U] {
	return &Future[U]{chain: func(ctx context.Context) (U, error) {
		v, err := f.Await(ctx)
		if err != nil {
			var zero U
			return zero, err
		}
		return fn(v)
	}}
}
