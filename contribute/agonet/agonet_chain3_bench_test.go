package agonet

// 链3 Peek 热路径基准（设计文档风险预案——拷贝化性能影响验证）：
//   BenchmarkPeekFrameHeader——帧头小 n（4 字节，每帧必经）拷贝化前后对比
// 用法：修复后 go test -bench=BenchmarkPeek -benchmem
//       修复前 git stash push contribute/agonet/connection.go 后同命令对比
import (
	"net"
	"testing"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/buffer/elastic"
	"github.com/valyala/bytebufferpool"
)

func BenchmarkPeekFrameHeader(b *testing.B) {
	c := &conn{rawConn: &net.TCPConn{}, buffer: bytebufferpool.Get(), inboundBuffer: elastic.RingBuffer{}}
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i)
	}
	_, _ = c.buffer.Write(payload) // 预装 1KB（inbound 空——走 Peek 路径①）
	defer bytebufferpool.Put(c.buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf, err := c.Peek(4) // 帧头 4 字节（小 n 热路径——每帧必经）
		if err != nil {
			b.Fatal(err)
		}
		_ = buf
	}
}

// BenchmarkFrameProcessing——解码器核心热路径（每帧序列：Peek 帧头 + Next 帧体）
// 设计预算判据：拷贝化对"帧处理端到端"的影响（纯 Peek 基线近乎零成本——+531% 是放大效应）
func BenchmarkFrameProcessing(b *testing.B) {
	c := &conn{rawConn: &net.TCPConn{}, buffer: bytebufferpool.Get(), inboundBuffer: elastic.RingBuffer{}}
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i)
	}
	_, _ = c.buffer.Write(payload)
	defer bytebufferpool.Put(c.buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 每帧：Peek(4) 帧头 → Next(1024) 帧体（真实解码器序列）
		hdr, err := c.Peek(4)
		if err != nil {
			b.Fatal(err)
		}
		_ = hdr[0]
		frame, err := c.Next(1024)
		if err != nil {
			b.Fatal(err)
		}
		_ = frame[0]
		// 注：Next 已全 make 副本（无池借出）——无需归还（byteslice.Put 已删——
		// 池化方案被 StickyPackets 否决后残留清理）
		// 消费复位：重新预装（模拟下包到达——每帧装新数据）
		c.buffer.Reset()
		if _, err := c.buffer.Write(payload); err != nil {
			b.Fatal(err)
		}
	}
}
