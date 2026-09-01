//go:build agonet_regress

package agonet

// 回归测试（white-box, package agonet）：F 根因组（协议/编解码正确性，根包部分）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> F1（commit 4d1491a4 快照）。
// TDD 红-绿语义：断言"正确行为"，缺陷未修复时【预期失败(红)】，修复后【通过(绿)】自动验证。
// 注意：mockNetConn/dummyAddr 桩定义在 agonet_regress_lifecycle_test.go（同包共享）。
// 运行：go test -race -tags agonet_regress . -run 'TestF1' -v

import (
	"testing"
)

// TestF1_NextZeroShouldReturnEmptyFrame 缺陷复现
//
// 期望行为：Next(0) 返回空帧 []byte{}，缓冲数据保留供后续粘包帧解码。
// 当前缺陷：Next(0) 在 n<=0 分支把 n 置为 totalLen（connection.go:138-140），
//           返回并消费整个缓冲 -> LengthFieldDecoder 遇空 body 帧时吞掉后续帧，流永久失步。
func TestF1_NextZeroShouldReturnEmptyFrame(t *testing.T) {
	c := newStreamConn(nil, mockNetConn{}, nil) // 测试即用即弃，conn 出作用域由 GC 回收

	payload := []byte{0x00, 0x03, 'A', 'B', 'C'} // 一个有效帧（后续粘包帧在真实场景紧跟其后）
	if _, err := c.buffer.Write(payload); err != nil {
		t.Fatal(err)
	}

	buf, err := c.Next(0)
	if err != nil {
		t.Fatalf("Next(0) err: %v", err)
	}

	if len(buf) != 0 {
		t.Fatalf("FAIL(F1 未修复): Next(0) 返回 %d 字节(期望空帧 0), 已消费缓冲并吞掉后续粘包帧; InboundBuffered=%d",
			len(buf), c.InboundBuffered())
	}
	if c.InboundBuffered() != len(payload) {
		t.Fatalf("FAIL(F1 未修复): Next(0) 后 InboundBuffered=%d, 期望 %d（数据被错误消费）",
			c.InboundBuffered(), len(payload))
	}
	t.Logf("PASS(F1 已修复): Next(0) 返回空帧, 缓冲保留 %d 字节", c.InboundBuffered())
}
