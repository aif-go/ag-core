package agonet

// 核心功能正确性测试（默认构建、绿测）：loadbalancer 调度分布。

import (
	"testing"
)

func TestRoundRobin_Distribution(t *testing.T) {
	lb := &roundRobinLoadBalancer{}
	const n = 4
	for i := 0; i < n; i++ {
		lb.register(&eventloop{idx: i})
	}

	counts := make(map[*eventloop]int)
	const rounds = 1000
	for i := 0; i < n*rounds; i++ {
		el := lb.next(nil)
		if el == nil {
			t.Fatal("next() returned nil with registered loops")
		}
		counts[el]++
	}
	for el, c := range counts {
		if c != rounds {
			t.Fatalf("round-robin imbalance: loop %d got %d, expect %d", el.idx, c, rounds)
		}
	}
}

func TestRoundRobin_Empty_ReturnsNil(t *testing.T) {
	// 未注册：返回 nil 而非除零 panic（G3 修复）。
	lb := &roundRobinLoadBalancer{}
	if el := lb.next(nil); el != nil {
		t.Fatalf("next() on empty lb=%v, expect nil", el)
	}
}

func TestRoundRobin_Size(t *testing.T) {
	lb := &roundRobinLoadBalancer{}
	lb.register(&eventloop{idx: 0})
	lb.register(&eventloop{idx: 1})
	if lb.len() != 2 {
		t.Fatalf("len()=%d, expect 2", lb.len())
	}
}

func TestLeastConnections_PicksMinConn(t *testing.T) {
	lb := &leastConnectionsLoadBalancer{}
	busy := &eventloop{idx: 0, connCount: 10}
	idle := &eventloop{idx: 1, connCount: 1}
	lb.register(busy)
	lb.register(idle)

	// 最小连接数应被选中。
	for i := 0; i < 10; i++ {
		if el := lb.next(nil); el != idle {
			t.Fatalf("least-conns picked loop %d (conn=%d), expect idle loop 1", el.idx, el.connCount)
		}
	}
}

func TestLeastConnections_Empty_ReturnsNil(t *testing.T) {
	lb := &leastConnectionsLoadBalancer{}
	if el := lb.next(nil); el != nil {
		t.Fatalf("next() on empty lb=%v, expect nil", el)
	}
}

func TestLeastConnections_FirstWhenEqual(t *testing.T) {
	// 连接数相同 → 取第一个。
	lb := &leastConnectionsLoadBalancer{}
	first := &eventloop{idx: 0, connCount: 3}
	second := &eventloop{idx: 1, connCount: 3}
	lb.register(first)
	lb.register(second)
	if el := lb.next(nil); el != first {
		t.Fatalf("equal conns must pick first, got loop %d", el.idx)
	}
}
