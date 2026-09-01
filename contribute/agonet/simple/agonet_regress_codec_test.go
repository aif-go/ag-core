//go:build agonet_regress

package simple

// 回归测试（white-box, package simple）：F 根因组（协议/编解码正确性，simple 包部分）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> F2（commit 4d1491a4 快照）。
// TDD 红-绿语义：断言"正确行为"，缺陷未修复时【预期失败(红)】，修复后【通过(绿)】自动验证。
// 正确行为：body 长度超出长度域容量时应 panic 报错（与字符串版 str_encoder 的 >fieldLen 检查一致），
//           而非静默截断长度域值。
// 运行：go test -race -tags agonet_regress ./simple/ -run TestF2 -v

import (
	"encoding/binary"
	"testing"
)

// captureOutboundCtx 最小 OutboundContext 桩：捕获 encoder 写出的最终字节。
type captureOutboundCtx struct {
	got []byte
}

func (c *captureOutboundCtx) Channel() Channel             { return nil }
func (c *captureOutboundCtx) Handler() Handler             { return nil }
func (c *captureOutboundCtx) Trigger(event any)            {}
func (c *captureOutboundCtx) Write(message any)            { c.got = message.([]byte) }
func (c *captureOutboundCtx) FireWrite(message any)        { c.Write(message) }
func (c *captureOutboundCtx) FireExceptionCaught(ex error) {}
func (c *captureOutboundCtx) FireEvent(event any)          {}

// TestF2_EncoderOverflowShouldError 正确行为断言（当前缺陷 -> 红）
//
// 期望行为：NewLengthFieldEncoder(fieldLen=2) 编码 body=70000（> 65535）时应 panic 报错
//           （长度超出长度域容量），而不是静默写出截断的长度域值 0x1170=4464。
// 当前缺陷：packFieldLength 用 uint16 静默截断，无任何报错 -> 对端按错误长度解析，帧错位/永久半包。
func TestF2_EncoderOverflowShouldError(t *testing.T) {
	enc := NewLengthFieldEncoder(binary.BigEndian, 2, 0, false).(OutboundHandler)
	body := make([]byte, 70000) // 超过 2 字节长度域容量(65535)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("FAIL(F2 未修复): body=%d 超出 fieldLen=2 容量却未报错, 长度域被静默截断(70000->4464), 对端将帧错位",
				len(body))
		}
		t.Logf("PASS(F2 已修复): 长度超出长度域容量时正确 panic: %v", r)
	}()

	ctx := &captureOutboundCtx{}
	enc.HandleWrite(ctx, body)
	// 未 panic -> 缺陷路径，由 defer 判定 FAIL
}
