package ag_cache

import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

// EngineFactoryParams 收集各引擎模块（如 agristretto.FxAgCacheRistrettoMode）
// 经 fx value group 注入的引擎工厂。
type EngineFactoryParams struct {
	fx.In
	Factories []EngineFactory `group:"agcache.engine"`
}

// NewAgCacheManager 从注入的引擎工厂与 core 装配配置构建 Manager：
// 填工厂 map，经 config 选择默认引擎，默认引擎未注册则快速失败。
func NewAgCacheManager(p EngineFactoryParams, props *AgCacheProperties) (*Manager, error) {
	m, err := NewManager(props)
	if err != nil {
		return nil, err
	}
	for _, f := range p.Factories {
		m.SetEngineFactory(f.Name(), f)
	}
	if f := m.EngineFactory(m.DefaultEngine()); f == nil {
		return nil, fmt.Errorf("agcache: engine %q not registered (modules: %v)", m.DefaultEngine(), factoryNames(p.Factories))
	}
	return m, nil
}

func factoryNames(fs []EngineFactory) []string {
	names := make([]string, 0, len(fs))
	for _, f := range fs {
		names = append(names, f.Name())
	}
	return names
}

// registerHooks 设置默认 Manager（使运行时 DefaultManager() 可用），
// 并注册 OnStop 钩子：仅 Close（v3 不跟踪统计）。
func registerHooks(lc fx.Lifecycle, m *Manager) {
	SetDefault(m)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return m.Close()
		},
	})
}

// FxAgCacheMode 是缓存 core 的自包含 Fx 模块：
// 收集所有引擎工厂（fx group）、建 Manager、装配 agcache.* 配置、
// 并接好生命周期钩子。
var FxAgCacheMode = fx.Module("ag_cache",
	fx.Provide(
		BindAgCacheProperties,
		NewAgCacheManager,
	),
	fx.Invoke(registerHooks),
)
