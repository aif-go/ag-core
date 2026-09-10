package agristretto

import (
	"errors"
	"testing"
	"time"

	"github.com/aif-go/ag-core/ag/ag_conf"
)

func newBinder(t *testing.T, source map[string]any) ag_conf.IBinder {
	t.Helper()
	ps := &ag_conf.MapPropertySource{}
	ps.Name = "test"
	ps.Source = source
	env, err := ag_conf.NewStandardEnvironment()
	if err != nil {
		t.Fatalf("NewStandardEnvironment: %v", err)
	}
	env.GetPropertySources().AddFirst(ps)
	return ag_conf.NewConfigurationPropertiesBinder(env)
}

// ──────── 绑定层叶子 + 容器编译断言 ────────

func TestConfig_LayerCompile(t *testing.T) {
	// RistrettoConfig（叶子，绑定层）存在且含四字段。
	var c RistrettoConfig
	_ = c.MaxCost
	_ = c.NumCounters
	_ = c.BufferItems
	_ = c.DefaultTTL
}

func TestConfigs_LayerCompile(t *testing.T) {
	// RistrettoConfigs（容器）含 Default + Namespaces。
	var cs RistrettoConfigs
	cs.Default.MaxCost = 1
	cs.Namespaces = map[string]RistrettoConfig{}
	if cs.Namespaces == nil {
		t.Fatal("Namespaces map should be usable")
	}
}

// ──────── 默认值 ────────

func TestDefaultRistrettoConfig(t *testing.T) {
	c := DefaultRistrettoConfig()
	if c.MaxCost != 100_000_000 {
		t.Fatalf("MaxCost = %d, want 100000000", c.MaxCost)
	}
	if c.NumCounters != 131_072 {
		t.Fatalf("NumCounters = %d, want 131072 (2^17)", c.NumCounters)
	}
	if c.BufferItems != 64 {
		t.Fatalf("BufferItems = %d, want 64", c.BufferItems)
	}
	if c.DefaultTTL != "0" {
		t.Fatalf("DefaultTTL = %q, want \"0\" (explicit never-expire)", c.DefaultTTL)
	}
}

// ──────── 绑定 ────────

func TestBindRistrettoConfig(t *testing.T) {
	binder := newBinder(t, map[string]any{
		"agcache.ristretto.default.maxcost":     "20971520",
		"agcache.ristretto.default.numcounters": "2097152",
		"agcache.ristretto.default.defaultttl":  "30s",
	})
	cfg, err := BindRistrettoConfig(binder)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if cfg.Default.MaxCost != 20971520 {
		t.Fatalf("Default.MaxCost = %d, want 20971520", cfg.Default.MaxCost)
	}
	if cfg.Default.NumCounters != 2097152 {
		t.Fatalf("Default.NumCounters = %d, want 2097152", cfg.Default.NumCounters)
	}
	if cfg.Default.DefaultTTL != "30s" {
		t.Fatalf("Default.DefaultTTL = %q, want 30s", cfg.Default.DefaultTTL)
	}
	if cfg.Namespaces == nil {
		t.Fatal("Namespaces should be initialized (empty map)")
	}
}

func TestBindRistrettoConfig_MissingKeepsDefault(t *testing.T) {
	binder := newBinder(t, map[string]any{})
	cfg, err := BindRistrettoConfig(binder)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if cfg.Default.MaxCost != 100_000_000 {
		t.Fatalf("Default.MaxCost = %d, want default 100000000", cfg.Default.MaxCost)
	}
	if cfg.Default.DefaultTTL != "0" {
		t.Fatalf("Default.DefaultTTL = %q, want \"0\" (default)", cfg.Default.DefaultTTL)
	}
}

// failBinder 是返回固定错误的 ag_conf.IBinder 测试替身。
type failBinder struct{}

func (failBinder) GetEnv() ag_conf.IConfigurableEnvironment { return nil }
func (failBinder) Bind(any, ...string) error                { return errors.New("bind failed") }

func TestBindRistrettoConfig_Error(t *testing.T) {
	if _, err := BindRistrettoConfig(failBinder{}); err == nil {
		t.Fatal("BindRistrettoConfig with failing binder should error")
	}
}

// ──────── parseTTL ────────

func TestParseTTL(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
		err  bool
	}{
		{"", 0, false},  // 未设置 → 默认 0（永不过期）
		{"0", 0, false}, // 永不过期
		{"60s", 60 * time.Second, false},
		{"60", 0, true},  // 纯数字 → 报错（防纳秒陷阱）
		{"abc", 0, true}, // 非法
	}
	for _, tt := range tests {
		got, err := parseTTL(tt.in)
		if (err != nil) != tt.err {
			t.Fatalf("parseTTL(%q) err = %v, want err=%v", tt.in, err, tt.err)
		}
		if !tt.err && got != tt.want {
			t.Fatalf("parseTTL(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

// ──────── ToOptions ────────

func TestToOptions(t *testing.T) {
	// 显式 TTL 转换。
	o, err := RistrettoConfig{DefaultTTL: "60s"}.ToOptions()
	if err != nil {
		t.Fatalf("ToOptions(60s): %v", err)
	}
	if o.DefaultTTL != 60*time.Second {
		t.Fatalf("DefaultTTL = %v, want 60s", o.DefaultTTL)
	}

	// 零值默认填充（固定默认，非 MaxCost×10% 推导）。
	o, err = RistrettoConfig{}.ToOptions()
	if err != nil {
		t.Fatalf("ToOptions(empty): %v", err)
	}
	if o.MaxCost != 100_000_000 || o.NumCounters != 131_072 || o.BufferItems != 64 {
		t.Fatalf("empty ToOptions defaults: MaxCost=%d NumCounters=%d BufferItems=%d, want 100MB/131072/64", o.MaxCost, o.NumCounters, o.BufferItems)
	}

	// 只给 MaxCost → NumCounters 保持固定默认（不随 MaxCost 推导）。
	o, err = RistrettoConfig{MaxCost: 200_000_000}.ToOptions()
	if err != nil {
		t.Fatalf("ToOptions(200MB): %v", err)
	}
	if o.MaxCost != 200_000_000 {
		t.Fatalf("MaxCost = %d, want 200000000", o.MaxCost)
	}
	if o.NumCounters != 131_072 {
		t.Fatalf("NumCounters = %d, want 131072 (fixed default, not derived)", o.NumCounters)
	}
}

func TestToOptions_NegativeRejected(t *testing.T) {
	if _, err := (RistrettoConfig{MaxCost: -1}).ToOptions(); err == nil {
		t.Fatal("negative MaxCost should error")
	}
	if _, err := (RistrettoConfig{NumCounters: -1}).ToOptions(); err == nil {
		t.Fatal("negative NumCounters should error")
	}
	if _, err := (RistrettoConfig{BufferItems: -1}).ToOptions(); err == nil {
		t.Fatal("negative BufferItems should error")
	}
}

func TestToOptions_InvalidTTLRejected(t *testing.T) {
	if _, err := (RistrettoConfig{DefaultTTL: "abc"}).ToOptions(); err == nil {
		t.Fatal("invalid DefaultTTL should error")
	}
	if _, err := (RistrettoConfig{DefaultTTL: "60"}).ToOptions(); err == nil {
		t.Fatal("pure-number DefaultTTL should error (nanosecond trap)")
	}
}

// ──────── mergeConfig（非零覆盖继承）───────

func TestMergeConfig_NonZeroOverrides(t *testing.T) {
	def := DefaultRistrettoConfig() // MaxCost=100MB, NumCounters=131072, BufferItems=64, TTL="0"
	nc := RistrettoConfig{NumCounters: 8_388_608}

	merged := mergeConfig(def, nc)
	if merged.NumCounters != 8_388_608 {
		t.Fatalf("NumCounters = %d, want 8388608 (override)", merged.NumCounters)
	}
	if merged.MaxCost != 100_000_000 {
		t.Fatalf("MaxCost = %d, want 100000000 (inherit)", merged.MaxCost)
	}
	if merged.BufferItems != 64 {
		t.Fatalf("BufferItems = %d, want 64 (inherit)", merged.BufferItems)
	}
	if merged.DefaultTTL != "0" {
		t.Fatalf("DefaultTTL = %q, want \"0\" (inherit)", merged.DefaultTTL)
	}
}
