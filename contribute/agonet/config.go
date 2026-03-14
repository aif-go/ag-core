package agonet

type ServerConfig struct {
	Address string

	Engine    EngineConfig    // 引擎配置
	KeepAlive KeepAliveConfig // 保持连接配置

	TLS  TLSConfig  // TLS配置
	TLCP TLCPConfig // TLCP配置
}

type TLSConfig struct {
}

type TLCPConfig struct {
}

type EngineConfig struct {
	NumEventLoop int  // 事件循环数量
	Multicore    bool // 是否多核心
	Ticker       bool // 是否使用ticker
}

type KeepAliveConfig struct {
	Enable   bool
	Idle     int // 空闲时间，单位秒
	Interval int // 间隔时间，单位秒
	Count    int
}
