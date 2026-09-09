package simple

import "testing"

// TestUnshareableHandler_ReusePanic 功能回归（Green，无 tag）：D2 实例共享强制。
//
// 语义：实现 UnshareableHandler 的 handler（如 idleHandler）只允许绑定一个 pipeline，
// 重复 AddLast 到第二个 pipeline 必须 panic（仿 Netty @Sharable 反义）。
func TestUnshareableHandler_ReusePanic(t *testing.T) {
	idle := IdleStateHandler(0, 0, 0, 0).(*idleHandler)

	p1 := NewPipeline()
	p1.AddLast(idle) // 第一次绑定：OK

	p2 := NewPipeline()
	defer func() {
		if recover() == nil {
			t.Fatal("复用不可共享 handler 到第二个 pipeline 应 panic")
		}
	}()
	p2.AddLast(idle) // 第二次绑定：必须 panic
}

// TestShareableHandler_ReuseOK 功能回归（Green，无 tag）：无状态 handler 默认可复用。
//
// 语义：未实现 UnshareableHandler 的 handler（SimpleInboundHandler 等）可安全复用
// 到多个 pipeline（short_client 模式），不受 D2 校验影响。
func TestShareableHandler_ReuseOK(t *testing.T) {
	h := NewSimpleInboundHandler(func(ctx InboundContext, msg []byte) {})

	p1 := NewPipeline()
	p2 := NewPipeline()
	p1.AddLast(h) // 可复用
	p2.AddLast(h) // 可复用（无 panic）
}
