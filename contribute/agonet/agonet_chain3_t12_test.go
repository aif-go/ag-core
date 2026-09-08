package agonet

// 链3 C4 修复测试（white-box，无 build tag 常驻）：
//   T12——Peek 别名跨连接池复用悬空（use-after-reuse）：
//     连接 A Peek 拿别名 → 关闭（内存还池）→ 连接 B 复用同一内存写不同数据
//     → 读 A 的别名——修复前内容被覆盖（红）/ 修复后副本独立（绿）
//   红态确定性：TestHzw1 实证（同 goroutine 内 Put 后紧接 Get 复用同一数组）
import (
	"net"
	"testing"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/buffer/elastic"
	"github.com/valyala/bytebufferpool"
)

func TestChain3_T12_PeekAliasSurvivesPoolReuse(t *testing.T) {
	// 池复用不保证（sync.Pool）——循环尝试直到命中复用（红态触发）或 N 次未命中
	// （行为锁定跳过——与 T3a/T8 同类）。bytebufferpool.Put 会 Reset（pool.go:76），
	// 复用对象 len=0——连接 B 写入从 [0] 覆盖别名区域。
	for attempt := 0; attempt < 8; attempt++ {
		// 连接 A（white-box 构造——真实 Peek 代码路径）
		cA := &conn{rawConn: &net.TCPConn{}, buffer: bytebufferpool.Get(), inboundBuffer: elastic.RingBuffer{}}
		if _, err := cA.buffer.Write([]byte("AAAAAAAAAA")); err != nil {
			t.Fatal(err)
		}
		peeked, err := cA.Peek(10) // 路径① 零拷贝别名（修复前）
		if err != nil {
			t.Fatal(err)
		}
		if string(peeked) != "AAAAAAAAAA" {
			t.Fatalf("初始读取错误: %q", peeked)
		}

		// "关闭"连接 A（模拟 release）：buffer 还池（Put 内部 Reset）
		bytebufferpool.Put(cA.buffer)
		cA.buffer = nil

		// 连接 B：尝试复用同一内存（同 goroutine Put 后 Get——sync.Pool 大概率复用）
		cB := &conn{rawConn: &net.TCPConn{}, buffer: bytebufferpool.Get(), inboundBuffer: elastic.RingBuffer{}}
		if _, err := cB.buffer.Write([]byte("BBBBBBBBBB")); err != nil {
			t.Fatal(err)
		}

		// 读连接 A 的别名：
		// 修复前（零拷贝别名）：指向已回池数组 → B 写入覆盖 → 脏（"BBBBBBBBBB"）→ 红态
		// 修复后（独立副本）：内容不变（"AAAAAAAAAA"）→ 绿
		if string(peeked) == "BBBBBBBBBB" {
			bytebufferpool.Put(cB.buffer)
			t.Fatalf("C4 实锤（attempt %d）：Peek 别名被跨连接池复用覆盖（use-after-reuse）——别名=%q", attempt, peeked)
		}
		if string(peeked) != "AAAAAAAAAA" {
			bytebufferpool.Put(cB.buffer)
			t.Fatalf("别名内容异常: %q（预期原内容 AAAAAAAAAA 或未覆盖）", peeked)
		}
		bytebufferpool.Put(cB.buffer)
	}
	// N 次未命中复用：池未复用同一内存——红态未触发（行为锁定：跳过断言）
	t.Log("SKIP: 池未复用同一内存（sync.Pool 不保证）——红态未触发，行为锁定")
}
