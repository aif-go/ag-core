package agristretto

import (
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

func TestDefaultRistrettoConfigProperties(t *testing.T) {
	p := DefaultRistrettoConfigProperties()
	if p.MaxCost != 100_000_000 {
		t.Fatalf("MaxCost = %d, want 100000000", p.MaxCost)
	}
	if p.NumCounters != 0 {
		t.Fatalf("NumCounters = %d, want 0 (derived)", p.NumCounters)
	}
}

func TestBindRistrettoConfigProperties(t *testing.T) {
	binder := newBinder(t, map[string]any{
		"agcache.ristretto.maxcost":     "20971520",
		"agcache.ristretto.numcounters": "2097152",
		"agcache.ristretto.defaultttl":  "30s",
	})
	props, err := BindRistrettoConfigProperties(binder)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if props.MaxCost != 20971520 {
		t.Fatalf("MaxCost = %d, want 20971520", props.MaxCost)
	}
	if props.NumCounters != 2097152 {
		t.Fatalf("NumCounters = %d, want 2097152", props.NumCounters)
	}
	if props.DefaultTTL != "30s" {
		t.Fatalf("DefaultTTL = %q, want 30s", props.DefaultTTL)
	}
}

func TestBindRistrettoConfigProperties_MissingKeepsDefault(t *testing.T) {
	binder := newBinder(t, map[string]any{})
	props, err := BindRistrettoConfigProperties(binder)
	if err != nil {
		t.Fatalf("bind: %v", err)
	}
	if props.MaxCost != 100_000_000 {
		t.Fatalf("MaxCost = %d, want default 100000000", props.MaxCost)
	}
	if props.DefaultTTL != "" {
		t.Fatalf("DefaultTTL = %q, want empty", props.DefaultTTL)
	}
}

func TestParseTTL(t *testing.T) {
	tests := []struct {
		in   string
		want time.Duration
		err  bool
	}{
		{"", 5 * time.Minute, false}, // 未设置 → 默认 5min（与 core 兜底一致）
		{"0", 0, false},              // 永不过期
		{"60s", 60 * time.Second, false},
		{"abc", 0, true},
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
