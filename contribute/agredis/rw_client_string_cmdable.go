package agredis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

/*
type StringCmdable interface {
	Append(ctx context.Context, key, value string) *IntCmd
	Decr(ctx context.Context, key string) *IntCmd
	DecrBy(ctx context.Context, key string, decrement int64) *IntCmd
	DelExArgs(ctx context.Context, key string, a DelExArgs) *IntCmd
	Digest(ctx context.Context, key string) *DigestCmd
	Get(ctx context.Context, key string) *StringCmd
	GetRange(ctx context.Context, key string, start, end int64) *StringCmd
	GetSet(ctx context.Context, key string, value interface{}) *StringCmd
	GetEx(ctx context.Context, key string, expiration time.Duration) *StringCmd
	GetDel(ctx context.Context, key string) *StringCmd
	Incr(ctx context.Context, key string) *IntCmd
	IncrBy(ctx context.Context, key string, value int64) *IntCmd
	IncrByFloat(ctx context.Context, key string, value float64) *FloatCmd
	LCS(ctx context.Context, q *LCSQuery) *LCSCmd
	MGet(ctx context.Context, keys ...string) *SliceCmd
	MSet(ctx context.Context, values ...interface{}) *StatusCmd
	MSetNX(ctx context.Context, values ...interface{}) *BoolCmd
	MSetEX(ctx context.Context, args MSetEXArgs, values ...interface{}) *IntCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *StatusCmd
	SetArgs(ctx context.Context, key string, value interface{}, a SetArgs) *StatusCmd
	SetEx(ctx context.Context, key string, value interface{}, expiration time.Duration) *StatusCmd
	SetIFEQ(ctx context.Context, key string, value interface{}, matchValue interface{}, expiration time.Duration) *StatusCmd
	SetIFEQGet(ctx context.Context, key string, value interface{}, matchValue interface{}, expiration time.Duration) *StringCmd
	SetIFNE(ctx context.Context, key string, value interface{}, matchValue interface{}, expiration time.Duration) *StatusCmd
	SetIFNEGet(ctx context.Context, key string, value interface{}, matchValue interface{}, expiration time.Duration) *StringCmd
	SetIFDEQ(ctx context.Context, key string, value interface{}, matchDigest uint64, expiration time.Duration) *StatusCmd
	SetIFDEQGet(ctx context.Context, key string, value interface{}, matchDigest uint64, expiration time.Duration) *StringCmd
	SetIFDNE(ctx context.Context, key string, value interface{}, matchDigest uint64, expiration time.Duration) *StatusCmd
	SetIFDNEGet(ctx context.Context, key string, value interface{}, matchDigest uint64, expiration time.Duration) *StringCmd
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *BoolCmd
	SetXX(ctx context.Context, key string, value interface{}, expiration time.Duration) *BoolCmd
	SetRange(ctx context.Context, key string, offset int64, value string) *IntCmd
	StrLen(ctx context.Context, key string) *IntCmd
}
*/

var _ redis.StringCmdable = (*RWClient)(nil)

// ====================== 读操作（使用从节点） ======================
func (c *RWClient) Get(ctx context.Context, key string) *redis.StringCmd {
	slave := c.getRandomSlave()
	return slave.Get(ctx, key)
}

func (c *RWClient) GetRange(ctx context.Context, key string, start, end int64) *redis.StringCmd {
	slave := c.getRandomSlave()
	return slave.GetRange(ctx, key, start, end)
}

func (c *RWClient) GetEx(ctx context.Context, key string, expiration time.Duration) *redis.StringCmd {
	slave := c.getRandomSlave()
	return slave.GetEx(ctx, key, expiration)
}

// GetDel 会删除key 属于写
// func (c *RWClient) GetDel(ctx context.Context, key string) *redis.StringCmd {
// 	slave := c.getRandomSlave()
// 	return slave.GetDel(ctx, key)
// }

func (c *RWClient) MGet(ctx context.Context, keys ...string) *redis.SliceCmd {
	slave := c.getRandomSlave()
	return slave.MGet(ctx, keys...)
}

func (c *RWClient) HGet(ctx context.Context, key, field string) *redis.StringCmd {
	slave := c.getRandomSlave()
	return slave.HGet(ctx, key, field)
}

func (c *RWClient) Exists(ctx context.Context, keys ...string) *redis.IntCmd {
	slave := c.getRandomSlave()
	return slave.Exists(ctx, keys...)
}

func (c *RWClient) StrLen(ctx context.Context, key string) *redis.IntCmd {
	slave := c.getRandomSlave()
	return slave.StrLen(ctx, key)
}
