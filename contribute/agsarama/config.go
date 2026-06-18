package agsarama

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// Config agsarama 配置结构体，提供对 sarama.Config 的配置友好封装
// 所有时间字段单位为毫秒，便于通过 YAML/JSON 配置文件进行序列化
type Config struct {
	Brokers []string

	Admin    ConfigAdmin    // Admin 相关配置
	Net      ConfigNet      // 网络相关配置
	Metadata ConfigMetadata // 元数据相关配置
	Producer ConfigProducer // 生产者相关配置
	Consumer ConfigConsumer // 消费者相关配置

	ClientID           string // 客户端标识符，用于日志和监控
	RackID             string // 机架标识符，用于机架感知
	ChannelBufferSize  int    // 通道缓冲区大小，默认 256
	ApiVersionsRequest bool   // 是否发送 ApiVersions 请求，默认 true
	Version            string // Kafka 版本，使用 sarama.ParseKafkaVersion 解析
}

// ConfigAdmin Admin 相关配置，对应 sarama.Config.Admin
type ConfigAdmin struct {
	Retry struct {
		Max     int   // 重试次数，默认 5
		Backoff int64 // 重试间隔，单位毫秒，默认 100ms
	}
	Timeout int64 // Admin 操作超时时间，单位毫秒，默认 3s
}

// ConfigNet 网络相关配置，对应 sarama.Config.Net
type ConfigNet struct {
	MaxOpenRequests int // 单个连接上允许的最大未完成请求数，默认 5

	DialTimeout  int64 // 建立连接超时时间，单位毫秒，默认 30s
	ReadTimeout  int64 // 读取响应超时时间，单位毫秒，默认 30s
	WriteTimeout int64 // 发送请求超时时间，单位毫秒，默认 30s

	ResolveCanonicalBootstrapServers bool // 是否解析bootstrap服务器的规范主机名，默认 false

	SASL struct {
		Enable       bool          // 是否启用 SASL 认证，默认 false
		Mechanism    SASLMechanism // SASL 机制，支持 plain, scram-sha-256, scram-sha-512, oauth, gssapi
		Version      int16         // SASL 协议版本，默认 1 (SASLHandshakeV1)
		Handshake    bool          // 是否发送 SASL 握手请求，默认 true
		AuthIdentity string        // 授权标识符，用于 SASL/PLAIN
		User         string        // 用户名，用于 SASL/PLAIN 和 SASL/SCRAM
		Password     string        // 密码，用于 SASL/PLAIN 和 SASL/SCRAM
		SCRAMAuthzID string        // SCRAM 授权标识符，用于 SASL/SCRAM
	}

	KeepAlive int64 // TCP keep-alive 时间，单位毫秒，默认 0 (禁用)
}

// ConfigMetadata 元数据相关配置，对应 sarama.Config.Metadata
type ConfigMetadata struct {
	Retry struct {
		Max     int   // 元数据请求重试次数，默认 3
		Backoff int64 // 重试间隔，单位毫秒，默认 250ms
	}
	RefreshFrequency       int64 // 元数据刷新频率，单位毫秒，默认 10分钟
	Full                   bool  // 是否维护完整的元数据，默认 true
	Timeout                int64 // 元数据请求超时时间，单位毫秒，默认 0 (禁用)
	AllowAutoTopicCreation bool  // 是否允许自动创建主题，默认 true
	SingleFlight           bool  // 是否使用 SingleFlight 模式，默认 true
}

// ConfigProducer 生产者相关配置，对应 sarama.Config.Producer
type ConfigProducer struct {
	MaxMessageBytes  int              // 单条消息最大字节数，默认 1MB
	RequiredAcks     RequiredAcks     // 确认机制，支持 no_response, wait_for_local, wait_for_all
	Timeout          int64            // 生产者请求超时时间，单位毫秒，默认 10s
	Compression      CompressionCodec // 压缩算法，支持 none, gzip, snappy, lz4, zstd
	CompressionLevel int              // 压缩级别，默认 -1 (使用默认压缩级别)
	Partitioner      PartitionerType  // 分区器构造函数，默认 HashPartitioner

	// 幂等生产者配置
	Idempotent  bool // 是否启用幂等生产者，默认 false
	Transaction struct {
		ID      string // 事务ID，用于事务性生产者
		Timeout int64  // 事务超时时间，单位毫秒，默认 1分钟

		Retry struct {
			Max     int   // 事务重试次数，默认 50
			Backoff int64 // 事务重试间隔，单位毫秒，默认 100ms
		}
	}

	// 返回通道配置
	Return struct {
		Successes bool // 是否返回成功消息通道，默认 false
		Errors    bool // 是否返回错误消息通道，默认 true
	}

	// 批量刷新配置
	Flush struct {
		Bytes       int   // 触发刷新的字节数阈值，默认 0
		Messages    int   // 触发刷新的消息数阈值，默认 0
		Frequency   int64 // 刷新频率，单位毫秒，默认 0
		MaxMessages int   // 单次刷新的最大消息数，默认 0
	}

	// 重试配置
	Retry struct {
		Max             int   // 发送消息重试次数，默认 3
		Backoff         int64 // 重试间隔，单位毫秒，默认 100ms
		MaxBufferLength int   // 重试缓冲区最大长度，默认 0 (无限制)
		MaxBufferBytes  int64 // 重试缓冲区最大字节数，默认 0 (无限制)
	}
}

// ConfigConsumer 消费者相关配置，对应 sarama.Config.Consumer
type ConfigConsumer struct {
	Group struct {
		Session struct {
			Timeout int64 // 会话超时时间，单位毫秒，默认 10s
		}
		Heartbeat struct {
			Interval int64 // 心跳间隔，单位毫秒，默认 3s
		}
		Rebalance struct {
			Timeout int64 // 重平衡超时时间，单位毫秒，默认 60s
			Retry   struct {
				Max     int   // 重平衡重试次数，默认 4
				Backoff int64 // 重平衡重试间隔，单位毫秒，默认 2s
			}
		}
		// Member struct {
		// 	UserData []byte // 用户自定义数据
		// }
		InstanceId          string // 消费者实例ID，用于静态成员资格
		ResetInvalidOffsets bool   // 是否重置无效偏移量，默认 true
	}
	Retry struct {
		Backoff int64 // 消费重试间隔，单位毫秒，默认 2s
	}
	Fetch struct {
		Min     int32 // 最小获取字节数，默认 1
		Default int32 // 默认获取字节数，默认 1MB
		Max     int32 // 最大获取字节数，默认 0 (无限制)
	}
	MaxWaitTime       int64 // 最大等待时间，单位毫秒，默认 500ms
	MaxProcessingTime int64 // 最大处理时间，单位毫秒，默认 100ms
	Return            struct {
		Errors bool // 是否返回错误通道，默认 false
	}
	Offsets struct {
		AutoCommit struct {
			Enable   bool  // 是否启用自动提交，默认 true
			Interval int64 // 自动提交间隔，单位毫秒，默认 1s
		}
		Initial   int64 // 初始偏移量，-1 表示 OffsetNewest，-2 表示 OffsetOldest
		Retention int64 // 偏移量保留时间，单位毫秒，默认 0 (禁用)
		Retry     struct {
			Max int // 偏移量提交重试次数，默认 3
		}
	}
	IsolationLevel IsolationLevel // 隔离级别，支持 read_uncommitted, read_committed
}

type RequiredAcks string

const (
	RequiredAcksNoResponse   RequiredAcks = "no_response"
	RequiredAcksWaitForLocal RequiredAcks = "wait_for_local"
	RequiredAcksWaitForAll   RequiredAcks = "wait_for_all"
)

func (r RequiredAcks) ToSarama() (sarama.RequiredAcks, error) {
	switch r {
	case RequiredAcksNoResponse:
		return sarama.NoResponse, nil
	case RequiredAcksWaitForLocal:
		return sarama.WaitForLocal, nil
	case RequiredAcksWaitForAll:
		return sarama.WaitForAll, nil
	default:
		return sarama.NoResponse, fmt.Errorf("invalid required acks: %s", r)
	}
}

type CompressionCodec string

const (
	CompressionNone   CompressionCodec = "none"
	CompressionGZIP   CompressionCodec = "gzip"
	CompressionSnappy CompressionCodec = "snappy"
	CompressionLZ4    CompressionCodec = "lz4"
	CompressionZSTD   CompressionCodec = "zstd"
)

func (c CompressionCodec) ToSarama() (sarama.CompressionCodec, error) {
	switch c {
	case CompressionNone:
		return sarama.CompressionNone, nil
	case CompressionGZIP:
		return sarama.CompressionGZIP, nil
	case CompressionSnappy:
		return sarama.CompressionSnappy, nil
	case CompressionLZ4:
		return sarama.CompressionLZ4, nil
	case CompressionZSTD:
		return sarama.CompressionZSTD, nil
	default:
		return sarama.CompressionNone, fmt.Errorf("invalid compression codec: %s", c)
	}
}

type SASLMechanism string

const (
	SASLTypePlaintext   SASLMechanism = "plain"
	SASLTypeSCRAMSHA256 SASLMechanism = "scram-sha-256"
	SASLTypeSCRAMSHA512 SASLMechanism = "scram-sha-512"
	SASLTypeOAuth       SASLMechanism = "oauth"
	SASLTypeGSSAPI      SASLMechanism = "gssapi"
)

func (m SASLMechanism) ToSarama() (sarama.SASLMechanism, error) {
	switch m {
	case SASLTypePlaintext:
		return sarama.SASLTypePlaintext, nil
	case SASLTypeSCRAMSHA256:
		return sarama.SASLTypeSCRAMSHA256, nil
	case SASLTypeSCRAMSHA512:
		return sarama.SASLTypeSCRAMSHA512, nil
	case SASLTypeOAuth:
		return sarama.SASLTypeOAuth, nil
	case SASLTypeGSSAPI:
		return sarama.SASLTypeGSSAPI, nil
	default:
		return sarama.SASLTypePlaintext, fmt.Errorf("invalid SASL mechanism: %s", m)
	}
}

type IsolationLevel string

const (
	IsolationLevelReadUncommitted IsolationLevel = "read_uncommitted"
	IsolationLevelReadCommitted   IsolationLevel = "read_committed"
)

func (i IsolationLevel) ToSarama() (sarama.IsolationLevel, error) {
	switch i {
	case IsolationLevelReadUncommitted:
		return sarama.ReadUncommitted, nil
	case IsolationLevelReadCommitted:
		return sarama.ReadCommitted, nil
	default:
		return sarama.ReadUncommitted, fmt.Errorf("invalid isolation level: %s", i)
	}
}

type PartitionerType string

const (
	PartitionerTypeHash         PartitionerType = "hash"
	PartitionerTypeManual       PartitionerType = "manual"
	PartitionerTypeRandom       PartitionerType = "random"
	PartitionerTypeRoundRobin   PartitionerType = "roundrobin"
)

// ToSarama converts the PartitionerType to a sarama.PartitionerConstructor.
// The matching is case-insensitive (via strings.ToLower), so "Manual", "manual",
// "MANUAL" etc. all resolve to the same partitioner.
func (p PartitionerType) ToSarama() (sarama.PartitionerConstructor, error) {
	switch strings.ToLower(string(p)) {
	case "hash":
		return sarama.NewHashPartitioner, nil
	case "manual":
		return sarama.NewManualPartitioner, nil
	case "random":
		return sarama.NewRandomPartitioner, nil
	case "roundrobin":
		return sarama.NewRoundRobinPartitioner, nil
	default:
		return nil, fmt.Errorf("agsarama: invalid partitioner type: %q (valid: hash, manual, random, roundrobin)", p)
	}
}

// NewDefaultConfig 返回带有默认值的配置
func NewDefaultConfig() *Config {
	c := &Config{}

	// Admin 默认值
	c.Admin.Retry.Max = 5
	c.Admin.Retry.Backoff = 100 // 100ms
	c.Admin.Timeout = 3000      // 3s

	// Net 默认值
	c.Net.MaxOpenRequests = 5
	c.Net.DialTimeout = 30000  // 30s
	c.Net.ReadTimeout = 30000  // 30s
	c.Net.WriteTimeout = 30000 // 30s
	c.Net.SASL.Handshake = true
	c.Net.SASL.Version = 1 // SASLHandshakeV1
	c.Net.KeepAlive = 0    // 禁用

	// Metadata 默认值
	c.Metadata.Retry.Max = 3
	c.Metadata.Retry.Backoff = 250       // 250ms
	c.Metadata.RefreshFrequency = 600000 // 10分钟
	c.Metadata.Full = true
	c.Metadata.AllowAutoTopicCreation = true
	c.Metadata.SingleFlight = true
	c.Metadata.Timeout = 0 // 禁用

	// Producer 默认值
	c.Producer.MaxMessageBytes = 1024 * 1024 // 1MB
	c.Producer.RequiredAcks = RequiredAcksWaitForLocal
	c.Producer.Timeout = 10000 // 10s
	c.Producer.Compression = CompressionNone
	c.Producer.CompressionLevel = -1 // 默认压缩级别
	c.Producer.Idempotent = false
	c.Producer.Transaction.Timeout = 60000 // 1分钟
	c.Producer.Transaction.Retry.Max = 50
	c.Producer.Transaction.Retry.Backoff = 100 // 100ms
	c.Producer.Return.Errors = true
	c.Producer.Return.Successes = false
	c.Producer.Flush.Bytes = 0
	c.Producer.Flush.Messages = 0
	c.Producer.Flush.Frequency = 0
	c.Producer.Flush.MaxMessages = 0
	c.Producer.Retry.Max = 3
	c.Producer.Retry.Backoff = 100 // 100ms
	c.Producer.Retry.MaxBufferLength = 0
	c.Producer.Retry.MaxBufferBytes = 0
	c.Producer.Partitioner = PartitionerTypeHash

	// Consumer 默认值
	c.Consumer.Group.Session.Timeout = 10000   // 10s
	c.Consumer.Group.Heartbeat.Interval = 3000 // 3s
	c.Consumer.Group.Rebalance.Timeout = 60000 // 60s
	c.Consumer.Group.Rebalance.Retry.Max = 4
	c.Consumer.Group.Rebalance.Retry.Backoff = 2000 // 2s
	c.Consumer.Group.ResetInvalidOffsets = true
	c.Consumer.Retry.Backoff = 2000 // 2s
	c.Consumer.Fetch.Min = 1
	c.Consumer.Fetch.Default = 1024 * 1024 // 1MB
	c.Consumer.Fetch.Max = 0               // 无限制
	c.Consumer.MaxWaitTime = 500           // 500ms
	c.Consumer.MaxProcessingTime = 100     // 100ms
	c.Consumer.Return.Errors = false
	c.Consumer.Offsets.AutoCommit.Enable = true
	c.Consumer.Offsets.AutoCommit.Interval = 1000 // 1s
	c.Consumer.Offsets.Initial = -1               // OffsetNewest
	c.Consumer.Offsets.Retention = 0              // 禁用
	c.Consumer.Offsets.Retry.Max = 3
	c.Consumer.IsolationLevel = IsolationLevelReadUncommitted

	// 全局默认值
	c.ClientID = "agsarama"
	c.ChannelBufferSize = 256
	c.ApiVersionsRequest = true
	// c.Version = "2.1.0" // 对应 sarama.V2_1_0_0
	c.Version = sarama.DefaultVersion.String()

	return c
}

// ToSaramaConfig 将 agsarama.Config 转换为 sarama.Config
func (c *Config) ToSaramaConfig() (*sarama.Config, error) {
	saramaConfig := sarama.NewConfig()

	// 转换 Admin 配置
	saramaConfig.Admin.Retry.Max = c.Admin.Retry.Max
	saramaConfig.Admin.Retry.Backoff = time.Duration(c.Admin.Retry.Backoff) * time.Millisecond
	saramaConfig.Admin.Timeout = time.Duration(c.Admin.Timeout) * time.Millisecond

	// 转换 Net 配置
	saramaConfig.Net.MaxOpenRequests = c.Net.MaxOpenRequests
	saramaConfig.Net.DialTimeout = time.Duration(c.Net.DialTimeout) * time.Millisecond
	saramaConfig.Net.ReadTimeout = time.Duration(c.Net.ReadTimeout) * time.Millisecond
	saramaConfig.Net.WriteTimeout = time.Duration(c.Net.WriteTimeout) * time.Millisecond
	saramaConfig.Net.ResolveCanonicalBootstrapServers = c.Net.ResolveCanonicalBootstrapServers
	saramaConfig.Net.KeepAlive = time.Duration(c.Net.KeepAlive) * time.Millisecond

	// 转换 SASL 配置
	saramaConfig.Net.SASL.Enable = c.Net.SASL.Enable
	if c.Net.SASL.Enable {
		mechanism, err := c.Net.SASL.Mechanism.ToSarama()
		if err != nil {
			return nil, err
		}
		saramaConfig.Net.SASL.Mechanism = mechanism
		saramaConfig.Net.SASL.Version = c.Net.SASL.Version
		saramaConfig.Net.SASL.Handshake = c.Net.SASL.Handshake
		saramaConfig.Net.SASL.AuthIdentity = c.Net.SASL.AuthIdentity
		saramaConfig.Net.SASL.User = c.Net.SASL.User
		saramaConfig.Net.SASL.Password = c.Net.SASL.Password
		saramaConfig.Net.SASL.SCRAMAuthzID = c.Net.SASL.SCRAMAuthzID
	}

	// 转换 Metadata 配置
	saramaConfig.Metadata.Retry.Max = c.Metadata.Retry.Max
	saramaConfig.Metadata.Retry.Backoff = time.Duration(c.Metadata.Retry.Backoff) * time.Millisecond
	saramaConfig.Metadata.RefreshFrequency = time.Duration(c.Metadata.RefreshFrequency) * time.Millisecond
	saramaConfig.Metadata.Full = c.Metadata.Full
	saramaConfig.Metadata.Timeout = time.Duration(c.Metadata.Timeout) * time.Millisecond
	saramaConfig.Metadata.AllowAutoTopicCreation = c.Metadata.AllowAutoTopicCreation
	saramaConfig.Metadata.SingleFlight = c.Metadata.SingleFlight

	// 转换 Producer 配置
	saramaConfig.Producer.MaxMessageBytes = c.Producer.MaxMessageBytes
	requiredAcks, err := c.Producer.RequiredAcks.ToSarama()
	if err != nil {
		return nil, err
	}
	saramaConfig.Producer.RequiredAcks = requiredAcks
	saramaConfig.Producer.Timeout = time.Duration(c.Producer.Timeout) * time.Millisecond
	partitioner, err := c.Producer.Partitioner.ToSarama()
	if err != nil {
		return nil, err
	} else if partitioner != nil {
		// 只有在指定了分区器时才设置，否则忽略配置，由sarama默认配置
		saramaConfig.Producer.Partitioner = partitioner
	}
	compression, err := c.Producer.Compression.ToSarama()
	if err != nil {
		return nil, err
	}
	saramaConfig.Producer.Compression = compression
	saramaConfig.Producer.CompressionLevel = c.Producer.CompressionLevel
	saramaConfig.Producer.Idempotent = c.Producer.Idempotent
	saramaConfig.Producer.Transaction.ID = c.Producer.Transaction.ID
	saramaConfig.Producer.Transaction.Timeout = time.Duration(c.Producer.Transaction.Timeout) * time.Millisecond
	saramaConfig.Producer.Transaction.Retry.Max = c.Producer.Transaction.Retry.Max
	saramaConfig.Producer.Transaction.Retry.Backoff = time.Duration(c.Producer.Transaction.Retry.Backoff) * time.Millisecond
	saramaConfig.Producer.Return.Successes = c.Producer.Return.Successes
	saramaConfig.Producer.Return.Errors = c.Producer.Return.Errors
	saramaConfig.Producer.Flush.Bytes = c.Producer.Flush.Bytes
	saramaConfig.Producer.Flush.Messages = c.Producer.Flush.Messages
	saramaConfig.Producer.Flush.Frequency = time.Duration(c.Producer.Flush.Frequency) * time.Millisecond
	saramaConfig.Producer.Flush.MaxMessages = c.Producer.Flush.MaxMessages
	saramaConfig.Producer.Retry.Max = c.Producer.Retry.Max
	saramaConfig.Producer.Retry.Backoff = time.Duration(c.Producer.Retry.Backoff) * time.Millisecond
	saramaConfig.Producer.Retry.MaxBufferLength = c.Producer.Retry.MaxBufferLength
	saramaConfig.Producer.Retry.MaxBufferBytes = c.Producer.Retry.MaxBufferBytes

	// 转换 Consumer 配置
	saramaConfig.Consumer.Group.Session.Timeout = time.Duration(c.Consumer.Group.Session.Timeout) * time.Millisecond
	saramaConfig.Consumer.Group.Heartbeat.Interval = time.Duration(c.Consumer.Group.Heartbeat.Interval) * time.Millisecond
	saramaConfig.Consumer.Group.Rebalance.Timeout = time.Duration(c.Consumer.Group.Rebalance.Timeout) * time.Millisecond
	saramaConfig.Consumer.Group.Rebalance.Retry.Max = c.Consumer.Group.Rebalance.Retry.Max
	saramaConfig.Consumer.Group.Rebalance.Retry.Backoff = time.Duration(c.Consumer.Group.Rebalance.Retry.Backoff) * time.Millisecond
	// saramaConfig.Consumer.Group.Member.UserData = c.Consumer.Group.Member.UserData
	saramaConfig.Consumer.Group.InstanceId = c.Consumer.Group.InstanceId
	saramaConfig.Consumer.Group.ResetInvalidOffsets = c.Consumer.Group.ResetInvalidOffsets
	saramaConfig.Consumer.Retry.Backoff = time.Duration(c.Consumer.Retry.Backoff) * time.Millisecond
	saramaConfig.Consumer.Fetch.Min = c.Consumer.Fetch.Min
	saramaConfig.Consumer.Fetch.Default = c.Consumer.Fetch.Default
	saramaConfig.Consumer.Fetch.Max = c.Consumer.Fetch.Max
	saramaConfig.Consumer.MaxWaitTime = time.Duration(c.Consumer.MaxWaitTime) * time.Millisecond
	saramaConfig.Consumer.MaxProcessingTime = time.Duration(c.Consumer.MaxProcessingTime) * time.Millisecond
	saramaConfig.Consumer.Return.Errors = c.Consumer.Return.Errors
	saramaConfig.Consumer.Offsets.AutoCommit.Enable = c.Consumer.Offsets.AutoCommit.Enable
	saramaConfig.Consumer.Offsets.AutoCommit.Interval = time.Duration(c.Consumer.Offsets.AutoCommit.Interval) * time.Millisecond
	saramaConfig.Consumer.Offsets.Initial = c.Consumer.Offsets.Initial
	saramaConfig.Consumer.Offsets.Retention = time.Duration(c.Consumer.Offsets.Retention) * time.Millisecond
	saramaConfig.Consumer.Offsets.Retry.Max = c.Consumer.Offsets.Retry.Max
	isolationLevel, err := c.Consumer.IsolationLevel.ToSarama()
	if err != nil {
		return nil, err
	}
	saramaConfig.Consumer.IsolationLevel = isolationLevel

	// 转换全局配置
	saramaConfig.ClientID = c.ClientID
	saramaConfig.RackID = c.RackID
	saramaConfig.ChannelBufferSize = c.ChannelBufferSize
	saramaConfig.ApiVersionsRequest = c.ApiVersionsRequest

	// 转换版本
	if c.Version != "" {
		version, err := sarama.ParseKafkaVersion(c.Version)
		if err != nil {
			return nil, fmt.Errorf("invalid kafka version: %s", c.Version)
		}
		saramaConfig.Version = version
	}

	return saramaConfig, nil
}

// Validate 验证配置的有效性
func (c *Config) Validate() error {
	// 这里可以添加自定义验证逻辑
	// 目前依赖 sarama.Config.Validate() 进行验证
	saramaConfig, err := c.ToSaramaConfig()
	if err != nil {
		return err
	}
	return saramaConfig.Validate()
}
