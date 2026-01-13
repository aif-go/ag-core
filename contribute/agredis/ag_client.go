package agredis

import (
	// "github.com/go-redis/redis/v8"
	"context"

	"github.com/redis/go-redis/v9"
)

var (
	_ AgRedisClient = (redis.UniversalClient)(nil)
	_ AgRedisClient = (*redis.Client)(nil)
	_ AgRedisClient = (*redis.ClusterClient)(nil)
	_ AgRedisClient = (*redis.Ring)(nil)
)

type AgRedisClient interface {
	// redis.UniversalClient // 通用客户端接口

	/* == Cmdable 子命令，只暴露必要的命令 == */
	// redis.Cmdable // 嵌入接口 最核心的嵌入接口 包含了所有 Redis 基础命令
	Echo(ctx context.Context, message interface{}) *redis.StringCmd
	Ping(ctx context.Context) *redis.StatusCmd

	// -- 内置核心功能 --
	redis.StringCmdable    // 操作 字符串（String） 类型
	redis.HashCmdable      // 操作 哈希表（Hash），适合存储对象（如用户资料）
	redis.ListCmdable      // 操作 列表（List）
	redis.SetCmdable       // 操作 集合（Set） —— 无序、唯一元素
	redis.SortedSetCmdable // 操作 有序集合（Sorted Set / ZSet）
	redis.BitMapCmdable    // 位图（Bitmap）操作
	redis.GenericCmdable   // 通用命令（不依赖特定数据类型）,所有 key 都能用的“元操作”
	// redis.HyperLogLogCmdable        // 基数统计相关命令
	// redis.GeoCmdable                // 地理空间（Geospatial）操作，基于 Redis 的地理位置索引（底层用 ZSet 实现）
	// redis.StreamCmdable             // 流相关命令，消息队列操作
	// redis.ScriptingFunctionsCmdable // 函数式脚本（Redis Functions，Redis 7+）,替代传统 Lua 脚本的新机制（Redis 7 引入）
	// redis.PubSubCmdable             // 发布/订阅（Publish-Subscribe）

	// -- 模块扩展功能(需要redis加载对应模块) --
	// redis.JSONCmdable          // 操作 JSON 文档（RedisJSON 模块）需要 Redis 加载 RedisJSON 模块
	// redis.SearchCmdable        // 全文搜索（RediSearch 模块）需要 Redis 加载 RediSearch 模块
	// redis.TimeseriesCmdable    // 时间序列（RedisTimeSeries 模块）
	// redis.ProbabilisticCmdable // 概率数据结构（RedisBloom 模块）
	// redis.VectorSetCmdable     // 向量相似度搜索（RedisVL / RedisAI 或 RediSearch 向量）

	// -- 运维/集群管理 --
	// redis.ACLCmdable     // 访问控制相关命令
	// redis.ClusterCmdable //Redis Cluster 管理命令

	AddHook(redis.Hook)                                                        // 给客户端添加「钩子函数」（比如监控命令执行、记录日志、统计耗时）
	Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error // 实现 Redis 事务的「乐观锁」（Watch 监听指定 Key，事务执行前若 Key 被修改则事务失败）

	// 高级接口不暴露
	// Do(ctx context.Context, args ...interface{}) *redis.Cmd                    // 执行自定义的 Redis 命令（比如 Redis 新增的命令未被 go-redis 封装时，用 Do 手动调用）
	// Process(ctx context.Context, cmd redis.Cmder) error                        // 手动处理 / 执行一个 Redis 命令对象（底层方法，业务层极少直接用）
	// Subscribe(ctx context.Context, channels ...string) *redis.PubSub           // 订阅普通频道
	// PSubscribe(ctx context.Context, channels ...string) *redis.PubSub          // 订阅模式匹配的频道
	// SSubscribe(ctx context.Context, channels ...string) *redis.PubSub          // 订阅 Redis Stream 频道

	Close() error                // 关闭客户端，释放连接池等资源（必须调用，否则会有连接泄露）
	PoolStats() *redis.PoolStats // 获取客户端连接池的统计信息（比如活跃连接数、空闲连接数、请求数等，用于监控）
}
