package usage_test

import (
	"context"
	"errors"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_cache/agristretto"
	"github.com/aif-go/ag-core/ag/ag_cache/test/usage"
	"github.com/aif-go/ag-core/ag/ag_conf"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestUsage_Complete 完整实践用法：
//  1. 真实 app.yml 配置加载（ag_conf reader → binder）
//  2. Fx 装配 core + ristretto 引擎
//  3. 业务 Service（Cache-Aside / 批量失效 / 监控探活）
//  4. 业务场景断言（读穿透命中、Clear 隔离、Del 失效重载、miss/后端语义）
func TestUsage_Complete(t *testing.T) {
	ctx := context.Background()

	// ── 1. 加载真实 app.yml（agcache.defaultEngine + agcache.ristretto.*）──
	env, err := ag_conf.NewStandardEnvironment()
	if err != nil {
		t.Fatalf("NewStandardEnvironment: %v", err)
	}
	if err := ag_conf.LoadConfigFile(env, "app.yml"); err != nil {
		t.Fatalf("LoadConfigFile(app.yml): %v", err)
	}
	binder := ag_conf.NewConfigurationPropertiesBinder(env)

	// ── 2. Fx 装配（core 收集引擎 group → Manager；引擎 Provide 到 group）──
	var mgr *ag_cache.Manager
	app := fxtest.New(t,
		fx.Provide(func() ag_conf.IBinder { return binder }),
		ag_cache.FxAgCacheMode,
		agristretto.FxAgCacheRistrettoMode,
		fx.Populate(&mgr), // 从 fx 图取出 *Manager
	)
	app.RequireStart()
	defer app.RequireStop()

	// ── 3. 业务对象（数据源 + Service；注入 Manager 构造时绑定缓存）──
	repo := &usage.UserRepo{DB: map[string]*usage.User{
		"u:1": {ID: "u:1", Name: "Alice"},
		"u:2": {ID: "u:2", Name: "Bob"},
	}}
	users := usage.NewUserService(mgr, repo)

	center := &usage.ParamCenter{Values: map[string]string{"host:port": "10.0.0.1:8080"}}
	params := usage.NewParamService(mgr, center)

	// ── 4. 业务场景 ──

	// 4.1 Cache-Aside：读用户（miss → loader → 缓存）
	u, err := users.GetUser(ctx, "u:1")
	if err != nil || u.Name != "Alice" {
		t.Fatalf("GetUser u:1: u=%+v err=%v", u, err)
	}
	// 探活确认已缓存命中（未再触发 loader）
	if name, ok := users.ProbeUser(ctx, "u:1"); !ok || name != "Alice" {
		t.Fatalf("ProbeUser u:1: name=%q ok=%v", name, ok)
	}

	// 4.2 参数缓存：读参数（读穿透）
	p, err := params.GetParam(ctx, "host:port")
	if err != nil || p.Value != "10.0.0.1:8080" {
		t.Fatalf("GetParam host:port: p=%+v err=%v", p, err)
	}

	// 4.3 参数更新 → 广播 Clear → 重载新值；users 不受影响
	center.Values["host:port"] = "10.0.0.2:8080" // 数据源更新
	if err := params.BroadcastUpdate(ctx); err != nil {
		t.Fatalf("BroadcastUpdate: %v", err)
	}
	p2, err := params.GetParam(ctx, "host:port")
	if err != nil || p2.Value != "10.0.0.2:8080" {
		t.Fatalf("GetParam after clear: p=%+v err=%v", p2, err)
	}
	if name, ok := users.ProbeUser(ctx, "u:1"); !ok || name != "Alice" {
		t.Fatalf("users should be unaffected by params.Clear: name=%q ok=%v", name, ok)
	}

	// 4.4 用户变更 → Del 失效 → 重载新值
	repo.DB["u:1"] = &usage.User{ID: "u:1", Name: "Alice2"}
	if err := users.RefreshUser(ctx, "u:1"); err != nil {
		t.Fatalf("RefreshUser: %v", err)
	}
	u2, err := users.GetUser(ctx, "u:1")
	if err != nil || u2.Name != "Alice2" {
		t.Fatalf("GetUser after refresh: u=%+v err=%v", u2, err)
	}

	// 4.5 纯读 miss 语义（Get → ErrCacheMiss）
	_, err = users.GetUser(ctx, "u:missing")
	if !errors.Is(err, ag_cache.ErrCacheMiss) {
		t.Fatalf("GetUser u:missing: want ErrCacheMiss, got %v", err)
	}
	// 探活 miss → ok=false，无 error
	if _, ok := users.ProbeUser(ctx, "u:missing"); ok {
		t.Fatal("ProbeUser u:missing should miss")
	}

	// 4.6 独立实例隔离：users 与 params 数据互不污染（u:2 独立加载）
	u3, err := users.GetUser(ctx, "u:2")
	if err != nil || u3.Name != "Bob" {
		t.Fatalf("GetUser u:2: u=%+v err=%v", u3, err)
	}
	// u:2 已缓存，params 的 Clear 不影响 users 的 u:2
	if name, ok := users.ProbeUser(ctx, "u:2"); !ok || name != "Bob" {
		t.Fatalf("users u:2 should be cached independently: name=%q ok=%v", name, ok)
	}
}
