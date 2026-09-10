package ag_cache_test

import (
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
	"github.com/aif-go/ag-core/ag/ag_cache/agristretto"
	"github.com/aif-go/ag-core/ag/ag_conf"

	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// newBinder 构建 ag_conf.IBinder，用扁平小写属性 map 填充，
// 模拟 YAML 展平键（agcache.defaultEngine、agcache.ristretto.maxCost, ...）。
func newBinder(src map[string]any) ag_conf.IBinder {
	ps := &ag_conf.MapPropertySource{}
	ps.Name = "test"
	ps.Source = src
	env, err := ag_conf.NewStandardEnvironment()
	if err != nil {
		panic(err)
	}
	env.GetPropertySources().AddFirst(ps)
	return ag_conf.NewConfigurationPropertiesBinder(env)
}

// startFx 用给定 YAML 属性 map 启动真实装配（core + agristretto 引擎），
// 返回 stop 函数。注册幂等，一个测试进程可启动多个 app。
func startFx(t *testing.T, yaml map[string]any) (stop func()) {
	t.Helper()
	binder := newBinder(yaml)
	app := fxtest.New(t,
		fx.Provide(func() ag_conf.IBinder { return binder }),
		ag_cache.FxAgCacheMode,
		agristretto.FxAgCacheRistrettoMode,
	)
	app.RequireStart()
	return func() {
		app.RequireStop()
		ag_cache.CloseAll() // 清默认 Manager，避免跨测试泄漏
	}
}

// dflt 返回默认 Manager（startFx 的 SetDefault 设置）。
func dflt() *ag_cache.Manager {
	m := ag_cache.DefaultManager()
	if m == nil {
		panic("no default manager")
	}
	return m
}

// mustCache 便捷获取 LoaderCache：测试中引擎正常，GetCacheWithLoader 的 err 应为 nil。
func mustCache[T any](t *testing.T, m *ag_cache.Manager, name string, loader ag_cache.LoaderFunc[T], opts ...ag_cache.Option[T]) *ag_cache.LoaderCache[T] {
	t.Helper()
	c, err := ag_cache.GetCacheWithLoader(m, name, loader, opts...)
	if err != nil {
		t.Fatalf("GetCacheWithLoader(%s): %v", name, err)
	}
	return c
}
