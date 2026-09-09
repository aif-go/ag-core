package agonet

// 验证：buf = B[:n] 共享原 cap——调用方 append 越界写污染 conn 缓冲
import (
	"fmt"
	"net"
	"testing"

	"github.com/aif-go/ag-core/contribute/agonet/pkg/buffer/elastic"
	"github.com/valyala/bytebufferpool"
)

func TestChain3_Proof_NextAliasAppendPollution(t *testing.T) {
	c := &conn{rawConn: &net.TCPConn{}, buffer: bytebufferpool.Get(), inboundBuffer: elastic.RingBuffer{}}

	// 包 1：10 字节
	c.buffer.Write([]byte("AAAAAAAAAA"))
	buf, err := c.Next(10) // 路径③：buf len=10，但 cap=原 cap（64 字节）
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("buf len=%d cap=%d（共享原 cap——调用方 append 会越界写）\n", len(buf), cap(buf))

	// 模拟调用方（解码器）贪心 append：
	buf = append(buf, 'X')
	fmt.Printf("调用方 append 后 buf=%q（写入 [10] 位置）\n", buf)

	// 模拟 el.read 收尾 + 下包到达（4 字节 ≤ 剩余 cap 6——append 不扩容，写同一旧数组）：
	c.buffer.Reset()
	c.buffer.Write([]byte("BBBB"))
	// 共享内存竞争：conn 写入 [10:14] 覆盖调用方 append 的 X——
	// 调用方视图 buf 的 [10] 现在是 conn 的下包数据（'B'），append 结果被破坏
	fmt.Printf("调用方视图 buf=%q（期望 \"AAAAAAAAAAX\"——若 [10] 变为 conn 的 B 则共享内存竞争实锤）\n", buf)
	if len(buf) > 10 && buf[10] == 'B' {
		t.Fatal("实锤：buf 共享原 cap——conn 下包写入覆盖了调用方 append 结果（双写竞争，双向数据损坏风险）")
	}
}
