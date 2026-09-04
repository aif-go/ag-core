package agonet

import "testing"

// TestInEventLoop_Semantics 功能回归（Green，无 tag）：G2 修复后 InEventLoop 语义保持正确。
//
// 修复背景：goroutineId 改 atomic.Int64 后，InEventLoop 仍须正确判断
// "当前 goroutine 是否为该 eventloop 的 goroutine"——非 eventloop goroutine 必须返回 false。
func TestInEventLoop_Semantics(t *testing.T) {
	el := &eventloop{}
	el.goroutineId.Store(int64(42)) // 模拟 run() 写入

	if el.InEventLoop() {
		t.Fatal("非 eventloop goroutine 调用 InEventLoop 应返回 false")
	}
}
