package agristretto

import (
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/dgraph-io/ristretto/v2"
)

// sketchDelta 测量创建单个 Ristretto 实例（NumCounters=nc）的堆增量。
// 关闭 GC 避免测量窗口内回收干扰；创建后回收实例再返回增量。
// 结果含实例固定开销（setBuf/bloom/goroutine）+ sketch，因此是"下界 + 固定底子"。
func sketchDelta(nc, maxCost int64) uint64 {
	runtime.GC()
	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)

	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	c, err := ristretto.NewCache[string, []byte](&ristretto.Config[string, []byte]{
		NumCounters: nc,
		MaxCost:     maxCost,
		BufferItems: 64,
	})
	if err != nil {
		debug.SetGCPercent(100)
		return 0
	}
	c.Get("keepalive") // 保持对象可达，防止被 GC
	runtime.ReadMemStats(&m1)
	c.Close()
	return m1.HeapAlloc - m0.HeapAlloc
}

const mb = 1 << 20

// TestSketch_AllocatedAtCreation 验证 sketch 在实例创建时即一次性预分配
// （heap 增量随 NumCounters 显著变化，而非按条目按需分配）。
func TestSketch_AllocatedAtCreation(t *testing.T) {
	small := sketchDelta(1_048_576, 100_000_000)   // 1M
	big := sketchDelta(16_777_216, 100_000_000)    // 16M
	bigger := sketchDelta(67_108_864, 100_000_000) // 64M

	t.Logf("NumCounters=1M  heap ~ %.1f MB", float64(small)/mb)
	t.Logf("NumCounters=16M heap ~ %.1f MB", float64(big)/mb)
	t.Logf("NumCounters=64M heap ~ %.1f MB", float64(bigger)/mb)

	// 创建即占：64M 实例应显著大于 1M 实例（sketch 2×n 线性），至少 8 倍。
	if bigger < small*8 {
		t.Fatalf("sketch should scale with NumCounters at creation: 1M=%.1fMB 64M=%.1fMB", float64(small)/mb, float64(bigger)/mb)
	}
}

// TestSketch_Next2PowerJump 验证 next2Power 取整代价：
// NumCounters=10M（非 2 的幂）会被取整放大，显著贵于相邻 2 的幂 8M。
// 8M→16M 线性翻倍（每 counter 固定字节），10M 应明显高于 8M 档。
func TestSketch_Next2PowerJump(t *testing.T) {
	eight := float64(sketchDelta(8_388_608, 100_000_000))    // 2^23
	ten := float64(sketchDelta(10_000_000, 100_000_000))     // 非 2 的幂
	sixteen := float64(sketchDelta(16_777_216, 100_000_000)) // 2^24

	t.Logf("NumCounters=8M  heap ~ %.1f MB", eight/mb)
	t.Logf("NumCounters=10M heap ~ %.1f MB", ten/mb)
	t.Logf("NumCounters=16M heap ~ %.1f MB", sixteen/mb)

	// 2 的幂档应近似线性翻倍：16M ≈ 2 × 8M（sketch 2B/counter 固定）。
	ratioLinear := sixteen / eight
	if ratioLinear < 1.7 || ratioLinear > 2.4 {
		t.Fatalf("2-power tiers should double: 16M/8M = %.2f (want ~2)", ratioLinear)
	}

	// 非 2 的幂（10M）应显著贵于相邻 8M 档（next2Power 取整放大）。
	ratioTenVsEight := ten / eight
	t.Logf("10M/8M = %.2f (8M 是 2 的幂，10M 非)", ratioTenVsEight)
	if ratioTenVsEight < 1.3 {
		t.Fatalf("non-power-of-2 10M should cost notably more than 8M: ratio=%.2f (want >1.3)", ratioTenVsEight)
	}
}

// TestSketch_PerInstanceIndependent 验证每实例独立预分配：
// N 个独立实例的总内存 ≈ N × 单实例（线性叠加）。
func TestSketch_PerInstanceIndependent(t *testing.T) {
	const nc = 2_097_152 // 2M

	single := sketchDelta(nc, 100_000_000)

	// 同时存活 3 个实例的总增量。
	runtime.GC()
	debug.SetGCPercent(-1)
	defer debug.SetGCPercent(100)
	var m0, m1 runtime.MemStats
	runtime.ReadMemStats(&m0)
	caches := make([]*ristretto.Cache[string, []byte], 0, 3)
	for i := 0; i < 3; i++ {
		c, err := ristretto.NewCache[string, []byte](&ristretto.Config[string, []byte]{
			NumCounters: nc, MaxCost: 100_000_000, BufferItems: 64,
		})
		if err != nil {
			t.Fatalf("NewCache: %v", err)
		}
		c.Get("keepalive")
		caches = append(caches, c)
	}
	runtime.ReadMemStats(&m1)
	for _, c := range caches {
		c.Close()
	}

	total := m1.HeapAlloc - m0.HeapAlloc
	perInst := float64(total) / mb / 3
	t.Logf("single@2M = %.1fMB, 3 instances total = %.1fMB (per-inst ~%.1fMB)", float64(single)/mb, float64(total)/mb, perInst)

	// 3 实例应为单实例的约 3 倍（每实例独立 sketch）。
	ratio := float64(total) / float64(single)
	if ratio < 2.2 || ratio > 3.8 {
		t.Fatalf("3 instances should be ~3x single instance memory: ratio=%.2f (want 2.2~3.8)", ratio)
	}
}
