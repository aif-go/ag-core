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
			Multicore:    true, // 默认多核心模式
			// Ticker:       false,
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
