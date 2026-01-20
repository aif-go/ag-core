package agredis

import (
	"ag-core/ag/ag_conf"
	"time"
)

const (
	AgRedisConfPrefix = "agredis"
)

type Type string

const (
	// TypeFailover        Type = "failover"
	// TypeFailoverCluster Type = "failovercluster"
	// TypeCluster         Type = "cluster"
	// TypeSingle          Type = "single"
	TypeUniversal Type = "universal"
	TypeRW        Type = "rw"
)

type AgRedisProperties struct {
	Type Type `value:"${:universal}"`

	Config AgUniversalOptionsProperties `json:",omitempty"`
	// 副本只在读写模式下生效，作为读节点使用
	Replicas []AgUniversalOptionsProperties `json:",omitempty"`
}

// AgUniversalOptionsProperties 通用配置
type AgUniversalOptionsProperties struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addrs []string `required:"true"`

	// ClientName will execute the `CLIENT SETNAME ClientName` command for each conn.
	ClientName string

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int

	// Common options.
	Protocol int
	Username string
	Password string

	// 连接 Sentinel 节点 时的认证信息
	SentinelUsername string
	SentinelPassword string

	MaxRetries      int
	MinRetryBackoff time.Duration // 单位：毫秒
	MaxRetryBackoff time.Duration // 单位：毫秒

	DialTimeout           time.Duration // 单位：毫秒
	ReadTimeout           time.Duration // 单位：毫秒
	WriteTimeout          time.Duration // 单位：毫秒
	ContextTimeoutEnabled bool

	// ReadBufferSize is the size of the bufio.Reader buffer for each connection.
	// Larger buffers can improve performance for commands that return large responses.
	// Smaller buffers can improve memory usage for larger pools.
	//
	// default: 32KiB (32768 bytes)
	ReadBufferSize int

	// WriteBufferSize is the size of the bufio.Writer buffer for each connection.
	// Larger buffers can improve performance for large pipelines and commands with many arguments.
	// Smaller buffers can improve memory usage for larger pools.
	//
	// default: 32KiB (32768 bytes)
	WriteBufferSize int

	// PoolFIFO uses FIFO mode for each node connection pool GET/PUT (default LIFO).
	PoolFIFO bool

	PoolSize        int
	PoolTimeout     time.Duration // 单位：毫秒
	MinIdleConns    int
	MaxIdleConns    int
	MaxActiveConns  int
	ConnMaxIdleTime time.Duration // 单位：毫秒
	ConnMaxLifetime time.Duration // 单位：毫秒

	// Only cluster clients.

	MaxRedirects   int
	ReadOnly       bool
	RouteByLatency bool
	RouteRandomly  bool

	// MasterName is the sentinel master name.
	// Only for failover clients.
	MasterName string

	// DisableIndentity - Disable set-lib on connect.
	//
	// default: false
	//
	// Deprecated: Use DisableIdentity instead.
	DisableIndentity bool

	// DisableIdentity is used to disable CLIENT SETINFO command on connect.
	//
	// default: false
	DisableIdentity bool

	IdentitySuffix string

	// FailingTimeoutSeconds is the timeout in seconds for marking a cluster node as failing.
	// When a node is marked as failing, it will be avoided for this duration.
	// Only applies to cluster clients. Default is 15 seconds.
	FailingTimeoutSeconds int

	UnstableResp3 bool

	// IsClusterMode can be used when only one Addrs is provided (e.g. Elasticache supports setting up cluster mode with configuration endpoint).
	IsClusterMode bool
}

// NewAgRedisPropertiesByBinder 从配置绑定器中创建 AgRedisProperties 实例
func NewAgRedisPropertiesByBinder(binder ag_conf.IBinder) (*AgRedisProperties, error) {
	conf := &AgRedisProperties{}
	err := binder.Bind(conf, AgRedisConfPrefix)
	if err != nil {
		return nil, err
	}
	return conf, nil
}
