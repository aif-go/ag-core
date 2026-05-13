package future

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/panjf2000/ants/v2"
)

// 最终结构
type Future[T any] struct {
	val   T
	err   error
	ready chan struct{} // 每次取出必须重新 make
	pool  *sync.Pool
	// ref      int32
	consumed int32 // 替代 ref：0=未消费, 1=已消费
}

// 按类型独立池
var poolMap sync.Map

func getPool[T any]() *sync.Pool {
	key := fmt.Sprintf("%T", *new(T))
	p, _ := poolMap.LoadOrStore(key, &sync.Pool{
		New: func() interface{} {
			return &Future[T]{} // 池里只存壳，chan 不预先创建
		},
	})
	return p.(*sync.Pool)
}

// -----------------------------------------------------------------------------
// 核心修复：
// 从池里拿出来后 → 必须 make(chan struct{})！
// -----------------------------------------------------------------------------
func NewFuture[T any](task func() (T, error)) *Future[T] {
	pool := getPool[T]()
	f := pool.Get().(*Future[T])

	// 🔥 关键：每次复用，必须重建 chan（唯一开销，极小）
	f.ready = make(chan struct{})

	// 重置其他字段
	f.val = *new(T)
	f.err = nil
	f.pool = pool

	// atomic.StoreInt32(&f.ref, 1)
	// consumed 池中可能残留1，需重置
	atomic.StoreInt32(&f.consumed, 0)

	// go func() {
	err := ants.Submit(func() {
		// panic 保护
		defer func() {
			if r := recover(); r != nil {
				f.err = fmt.Errorf("panic: %v", r)
			}
			close(f.ready) // 关闭本次的 chan
		}()

		// 执行任务
		val, err := task()
		f.val = val
		f.err = err
	})
	if err != nil {
		f.err = err
		close(f.ready)
	}

	return f
}

// Await 不变
func (f *Future[T]) Await(ctx context.Context) (T, error) {
	// CAS: 0→1，只允许一个调用者进入
	if !atomic.CompareAndSwapInt32(&f.consumed, 0, 1) {
		var zero T
		return zero, fmt.Errorf("future already consumed")
	}

	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case <-f.ready:
	}

	val := f.val
	err := f.err

	// // 最后一个调用者放回池,解决同一个future异步并发Await的问题
	// if atomic.AddInt32(&f.ref, -1) == 0 {
	f.pool.Put(f)
	// }

	return val, err
}
