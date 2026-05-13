package goroutine

import (
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
)

const (
	// DefaultAntsPoolSize sets up the capacity of worker pool, 256 * 1024.
	DefaultAntsPoolSize = 1 << 18

	// ExpiryDuration is the interval time to clean up those expired workers.
	ExpiryDuration = 10 * time.Second

	// Nonblocking decides what to do when submitting a new task to a full worker pool: waiting for a available worker
	// or returning nil directly.
	Nonblocking = true
)

func init() {
	// It releases the default pool from ants.
	// ants.Release() // FIXME houzw 不要默认释放默认池,会对其他使用ants框架的应用逻辑产生误解
}

// DefaultWorkerPool is the global worker pool.
var DefaultWorkerPool = Default()

// Pool is the alias of ants.Pool.
type Pool = ants.Pool

// Default instantiates a non-blocking goroutine pool with the capacity of DefaultAntsPoolSize.
func Default() *Pool {
	options := ants.Options{
		ExpiryDuration: ExpiryDuration,
		Nonblocking:    Nonblocking,
		PanicHandler: func(a any) {
			slog.Error("goroutine pool panic", "err", a)
		},
	}
	defaultAntsPool, _ := ants.NewPool(DefaultAntsPoolSize, ants.WithOptions(options))
	return defaultAntsPool
}
