package agonet

// 链1 P1 T5：send/trySend 单测（期望 API 驱动，独立 build tag）。
// 当前（修复前）：el.send / el.trySend 不存在 → 本文件编译失败（TDD 编译红，驱动 API 出现）。
// 修复后（R2）：编译通过 + 断言绿。
//
// 对应详细设计：[[agonet-链1修复详细设计]] §三 R2 / §五 T5。

import (
	"context"
	"testing"

	"golang.org/x/sync/errgroup"
)

func TestChain1_T5_SendTrySend(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	eng := &engine{
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{ctx: ctx},
	}
	el := &eventloop{ch: make(chan any, 4), eng: eng}

	// ① send 成功（ch 有空位）
	if !el.send(struct{}{}) {
		t.Fatal("send should succeed when ch has space")
	}

	// ② 填满 ch（已有 1 个，再塞 3 个 = 4 满）→ trySend 失败
	for i := 0; i < 3; i++ {
		el.ch <- struct{}{}
	}
	if el.trySend(struct{}{}) {
		t.Fatal("trySend should fail when ch full")
	}

	// ③ 引擎关闭 → send 返回 false
	cancel()
	if el.send(struct{}{}) {
		t.Fatal("send should fail after engine shutdown")
	}

	// ④ 新 el（ch 空）+ 引擎关闭 → send 仍 false（引擎关闭优先于 ch 空间）
	ctx2, cancel2 := context.WithCancel(context.Background())
	eng2 := &engine{
		concurrency: struct {
			*errgroup.Group
			ctx context.Context
		}{ctx: ctx2},
	}
	el2 := &eventloop{ch: make(chan any, 4), eng: eng2}
	cancel2()
	if el2.send(struct{}{}) {
		t.Fatal("send should fail after engine shutdown even when ch has space")
	}
}
