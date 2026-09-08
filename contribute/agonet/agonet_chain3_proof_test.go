package agonet

// C4 最小化技术验证（不启动 server/client——直接构造 conn，走真实 Next/Peek 代码路径，
// 模拟 el.read 的 buffer.Reset + 下包 unpackTCPConn 写入，验证零拷贝别名被覆盖）。
import (
	"fmt"
	"net"
	"testing"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/buffer/elastic"
	"github.com/valyala/bytebufferpool"
)

// newProofConn 构造最小 conn（rawConn 仅作非 nil 检查——Next/Peek 不调用它）。
func newProofConn() *conn {
	return &conn{
		rawConn:       &net.TCPConn{},
		buffer:        bytebufferpool.Get(),
		inboundBuffer: elastic.RingBuffer{},
	}
}

func TestChain3_Proof_NextAliasOverwritten(t *testing.T) {
	c := newProofConn()

	// 包 1 到达（模拟 unpackTCPConn 写入）
	c.buffer.Write([]byte("AAAAAAAAAA"))
	buf, err := c.Next(10)
	if err != nil {
		t.Fatal(err)
	}
	if string(buf) != "AAAAAAAAAA" {
		t.Fatalf("初始读取错误: %q", buf)
	}

	// el.read 收尾（OnTraffic 返回后）：buffer.Reset
	c.buffer.Reset()
	// 下包到达（模拟 unpackTCPConn 写入包 2）
	c.buffer.Write([]byte("BBBBBBBBBB"))

	// 解码器此刻读 buf（跨事件持有）——内容应为 A（包 1），实际?
	fmt.Printf("Next 别名内容: %q（期望 AAAAAAAAAA——若为 BBBBBBBBBB 则 C4 实锤：别名被下包覆盖）\n", buf)
	if string(buf) == "BBBBBBBBBB" {
		t.Fatal("C4 实锤：Next 路径③ 零拷贝别名被下包数据覆盖（use-after-reuse）")
	}
}

func TestChain3_Proof_PeekAliasOverwritten(t *testing.T) {
	c := newProofConn()

	// 包 1 到达（模拟 unpackTCPConn 写入）
	c.buffer.Write([]byte("AAAAAAAAAA"))
	peeked, err := c.Peek(10)
	if err != nil {
		t.Fatal(err)
	}

	// el.read 收尾 + 下包到达（同上）
	c.buffer.Reset()
	c.buffer.Write([]byte("BBBBBBBBBB"))

	fmt.Printf("Peek 别名内容: %q（期望 AAAAAAAAAA——若为 BBBBBBBBBB 则 C4 实锤）\n", peeked)
	if string(peeked) == "BBBBBBBBBB" {
		t.Fatal("C4 实锤：Peek 路径① 零拷贝别名被下包数据覆盖（use-after-reuse）")
	}
}
