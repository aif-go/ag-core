package simple

import "sync/atomic"

// UnshareableHandler 有状态 handler 的标记接口：声明"不可跨 pipeline 复用"。
//
// 语义：pipeline.AddFirst/AddLast 时若 handler 实现了本接口，则调用 TryBind 校验——
// 已绑定其他 pipeline 的实例再次入链会 panic（仿 Netty @Sharable 反义）。
// 有状态 handler（如 idleHandler）应嵌入 UnshareableHandlerBase 开箱即用；
// 无状态 handler 不实现本接口，默认可复用（不影响现有用法）。
type UnshareableHandler interface {
	// TryBind 尝试绑定当前 pipeline；返回 false = 已被其他 pipeline 绑定。
	TryBind() bool
}

// UnshareableHandlerBase UnshareableHandler 的默认实现。
// 嵌入即可获得"不可共享"能力（bound 状态封装在内部，pipeline 通过 TryBind 访问）。
type UnshareableHandlerBase struct {
	bound atomic.Bool
}

// TryBind impl UnshareableHandler：CAS 置位，仅第一次调用返回 true。
func (b *UnshareableHandlerBase) TryBind() bool {
	return b.bound.CompareAndSwap(false, true)
}
