package future

import "context"

// Future 用 channel 模拟
type Future[T any] struct {
	ch chan struct {
		val T
		err error
	}
}

func NewFuture[T any](f func() (T, error)) *Future[T] {
	fut := &Future[T]{ch: make(chan struct {
		val T
		err error
	}, 1)}
	go func() {
		val, err := f()
		fut.ch <- struct {
			val T
			err error
		}{val, err}
	}()
	return fut
}

func (f *Future[T]) Await(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case res := <-f.ch:
		return res.val, res.err
	}
}
