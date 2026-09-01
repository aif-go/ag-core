//go:build agonet_regress

package agonet

// 回归测试（white-box, package agonet）：并发竞态组（G1/G2，根包部分）。
// 对应跟踪清单：agonet-问题总报告与跟踪.md -> G1 / G2（commit 4d1491a4 快照）。
// 语义：期望"无 data race"。修复前 `go test -race -tags agonet_regress` 下被 race detector 判定 FAIL(红)；
//       修复后（atomic）无 race 即 PASS(绿)。必须 -race 运行才有效。
// 运行：go test -race -tags agonet_regress . -run 'TestG1|TestG2' -v

import (
	"sync"
	"testing"
)

// TestG1_LoadBalancerNextShouldBeAtomic 缺陷复现
//
// 缺陷：roundRobinLoadBalancer.next 的 lb.nextIndex++（load_balancer.go:71）无锁无 atomic，
//       多 listener 多 Accept goroutine 并发调用（engine.go:108-112,145）-> data race。
// 期望：nextIndex 使用 atomic，并发 next() 无 race。
func TestG1_LoadBalancerNextShouldBeAtomic(t *testing.T) {
	lb := &roundRobinLoadBalancer{}
	for i := 0; i < 4; i++ {
		lb.register(&eventloop{idx: i})
	}

	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50000; i++ {
				_ = lb.next(nil)
			}
		}()
	}
	wg.Wait()

	t.Log("PASS(G1 已修复): 并发 next() 无 data race（当前 -race 下若出现 race 则为红）")
}

// TestG2_EventLoopGoroutineIDShouldBeAtomic 缺陷复现
//
// 缺陷：eventloop.goroutineId 在 run()（event_loop.go:44）写入，InEventLoop()（event_loop.go:199-203）
//       由外部 goroutine 读取（channel.Write1 -> channel.go:94），无同步屏障 -> data race。
// 期望：goroutineId 使用 atomic，跨 goroutine 读写无 race。
func TestG2_EventLoopGoroutineIDShouldBeAtomic(t *testing.T) {
	el := &eventloop{}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // 模拟 run() 写 goroutineId
		defer wg.Done()
		for i := 0; i < 50000; i++ {
			el.goroutineId = int64(i)
		}
	}()

	wg.Add(1)
	go func() { // 模拟外部 goroutine 调 InEventLoop 读 goroutineId
		defer wg.Done()
		for i := 0; i < 50000; i++ {
			_ = el.InEventLoop()
		}
	}()
	wg.Wait()

	t.Log("PASS(G2 已修复): goroutineId 跨 goroutine 读写无 data race（当前 -race 下若出现 race 则为红）")
}
