//go:build agonet_regress

package agonet

// 回归测试（white-box, package agonet）：G/C 根因组（生命周期/连接所有权）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> G3 / G4 / C6（commit 4d1491a4 快照）。
// TDD 红-绿语义：断言"正确行为"，缺陷未修复时【预期失败(红)】，修复后【通过(绿)】自动验证。
// 本文件同时定义公共测试桩 mockNetConn/dummyAddr，供同包 F1（agonet_regress_codec_test.go）共享。
// 运行：go test -race -tags agonet_regress . -run 'TestG3|TestG4|TestC6' -v

import (
	"net"
	"testing"
)

// ---- 测试桩：最小 net.Conn 实现（newStreamConn 需要 LocalAddr/RemoteAddr 非 nil） ----
// 供本包 F1（Next(0)）与 C6（release 后 Read）共享。

type dummyAddr struct{}

func (dummyAddr) Network() string { return "mock" }
func (dummyAddr) String() string  { return "mock" }

type mockNetConn struct {
	net.Conn
}

func (m mockNetConn) LocalAddr() net.Addr  { return dummyAddr{} }
func (m mockNetConn) RemoteAddr() net.Addr { return dummyAddr{} }

// TestG3_DialBeforeStartShouldNotPanic 缺陷复现
//
// 期望行为：返回可处理的错误（如 ErrEngineInShutdown / ErrInvalidNetConn），绝不 panic。
// 当前缺陷：leastConnectionsLoadBalancer.next 对空 eventLoops 取 [0] -> index out of range。
func TestG3_DialBeforeStartShouldNotPanic(t *testing.T) {
	cfg := DefaultClientConfig()
	cli, err := NewClient(&BuiltinEventEngine{}, &cfg)
	if err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() { // 保证 net.Dial 成功，进入 EnrollContext -> eventLoops.next(nil)
		for {
			c, e := ln.Accept()
			if e != nil {
				return
			}
			_ = c.Close()
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FAIL(G3 未修复): 未 Start 即 Dial panic: %v（期望返回错误而非 panic）", r)
		}
	}()

	_, err = cli.Dial("tcp", ln.Addr().String())
	if err == nil {
		t.Fatal("FAIL(G3 部分修复?): Dial 返回 nil error, 但引擎未启动不应成功")
	}
	t.Logf("PASS(G3 已修复): 未 Start 即 Dial 返回错误而非 panic: %v", err)
}

// TestG4_StopBeforeStartShouldNotPanic 缺陷复现
//
// 期望行为：Stop 安全返回（幂等/空操作），绝不 panic。
// 当前缺陷：server.Stop 直接 s.eng.shutdown(nil)，而 s.eng 仅在 run() 中赋值 -> nil deref。
func TestG4_StopBeforeStartShouldNotPanic(t *testing.T) {
	s, err := NewServer(&BuiltinEventEngine{}, &ServerConfig{Addr: "tcp://127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FAIL(G4 未修复): 未 Start 即 Stop panic: %v（期望安全返回）", r)
		}
	}()

	if err := s.Stop(); err != nil {
		t.Fatalf("Stop 返回错误: %v", err)
	}
	t.Log("PASS(G4 已修复): 未 Start 即 Stop 安全返回")
}

// TestC6_ConnReadAfterReleaseShouldReturnClosed 缺陷复现
//
// 缺陷：el.close 调 c.release()（connection.go:372-392）把 rawConn/buffer 置 nil、inboundBuffer.Done() 归还池，
//       而 conn.Read（connection.go:97-114）在 release 后 c.buffer==nil 时 `c.buffer.B` nil 解引用 panic；
//       Next/Peek/WriteTo 同样无 ErrClosed 守卫（对比 Write 有 rawConn==nil 检查）。
// 期望：release 后 Read 返回 net.ErrClosed（或等效），绝不 panic。
func TestC6_ConnReadAfterReleaseShouldReturnClosed(t *testing.T) {
	c := newStreamConn(nil, mockNetConn{}, nil)
	_, _ = c.buffer.Write([]byte{1, 2, 3})
	c.release() // 归还缓冲池, rawConn/buffer 置 nil

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FAIL(C6 未修复): release 后 Read panic: %v（期望返回 net.ErrClosed 而非 panic）", r)
		}
	}()

	buf := make([]byte, 4)
	n, err := c.Read(buf)
	if err == nil {
		t.Fatalf("FAIL(C6 未修复): release 后 Read 返回 nil error (n=%d), 期望 net.ErrClosed", n)
	}
	t.Logf("PASS(C6 已修复): release 后 Read 返回 %v", err)
}
