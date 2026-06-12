package simple

import (
	"github.com/aif-go/ag-core/contribute/agonet/pkg/aerrors"
	"fmt"
	"sync"
	"time"
)

type Promise interface {
	Resolve(val any)
	Reject(err error)
	Await() (any, error)
	AwaitTimeout(timeout time.Duration) (any, error)
}

// 1. 定义 simplePromise 结构体
type simplePromise struct {
	done chan struct{} // 完成信号
	val  any           // 成功结果
	err  error         // 失败错误
	once sync.Once     // 保证只执行一次完成
}

// NewPromise 创建一个 Promise
func NewPromise() Promise {
	return &simplePromise{
		done: make(chan struct{}),
	}
}

// Resolve 成功完成（设置结果）
func (p *simplePromise) Resolve(val interface{}) {
	p.once.Do(func() {
		p.val = val
		close(p.done) // 关闭通道 = 通知完成
	})
}

// Reject 失败完成（设置错误）
func (p *simplePromise) Reject(err error) {
	p.once.Do(func() {
		p.err = err
		close(p.done)
	})
}

// Await 同步阻塞等待结果
func (p *simplePromise) Await() (any, error) {
	<-p.done // 阻塞直到完成
	return p.val, p.err
}

// AwaitTimeout 带超时的阻塞等待
func (p *simplePromise) AwaitTimeout(timeout time.Duration) (any, error) {
	select {
	case <-p.done:
		return p.val, p.err
	case <-time.After(timeout):
		return nil, fmt.Errorf("timeout: %v, %w", timeout, aerrors.ErrPromiseTimeout)
	}
}
