package agonet

type ServerConfig struct {
	Addr string

	Config OptionsConfig
}

type ClientConfig struct {
	Config OptionsConfig
}

// OptionsConfig 客户端、服务端通用配置
type OptionsConfig struct {
	Engine    EngineConfig    // 引擎配置
	KeepAlive KeepAliveConfig // 保持连接配置

	Security SecurityConfig // 安全配置
}

type EngineConfig struct {
	NumEventLoop int  // 事件循环数量
	Multicore    bool // 是否多核心
	// Ticker       bool // 是否使用ticker

	// ShutdownTimeout 优雅关闭 drain 超时（秒，0=默认 5s）：引擎关闭时在途事件处理上限
	ShutdownTimeout int
	// MaxConn 连接上限（0=不限制）：超限连接静默拒绝（应用层配额，非精确边界）
	MaxConn int32
}

type KeepAliveConfig struct {
	Enable   bool
	Idle     int // 空闲时间，单位秒
	Interval int // 间隔时间，单位秒
	Count    int
}

func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		Addr:   "tcp://:9000",
		Config: DefaultCommonConfig(),
	}
}

func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		Config: DefaultCommonConfig(),
	}
}

func DefaultCommonConfig() OptionsConfig {
	return OptionsConfig{
		Engine: EngineConfig{
			NumEventLoop: 0,
			Multicore:    true,
			// Ticker:       false,
			// 优雅关闭 drain 超时（秒）：显式 5s 与 engine.shutdownTimeout 兜底一致（0 同样回落到 5s）
			ShutdownTimeout: 5,
			// 连接上限：0 = 不限制（默认）
			MaxConn: 0,
		},
		KeepAlive: KeepAliveConfig{
			Enable:   true,
			Idle:     60, // 空闲时间，单位秒，默认60秒
			Interval: 12,
			Count:    5,
		},
		Security: DefaultSecurityConfig(),
	}
}
