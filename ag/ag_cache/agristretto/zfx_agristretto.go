package agristretto

import (
	"go.uber.org/fx"
)

// FxAgCacheRistrettoMode 将 Ristretto 引擎工厂贡献给 core 的
// "agcache.engine" fx group。依赖图自动装配：
//
//	IBinder ──BindRistrettoConfig──▶ *RistrettoConfigs ──NewAgristrettoFactory──▶ EngineFactory
//
// NewAgristrettoFactory 在装配期预解析校验全部配置（Default + 所有 Namespaces），
// 任一非法 → 构造返回 error → fx 启动失败（配置错误启动即死，非运行时 panic）。
var FxAgCacheRistrettoMode = fx.Module("ag_cache.agristretto",
	fx.Provide(
		BindRistrettoConfig, // 职责1：绑定容器
		fx.Annotate(NewAgristrettoFactory, fx.ResultTags(`group:"agcache.engine"`)), // 职责2：构造+预解析
	),
)
