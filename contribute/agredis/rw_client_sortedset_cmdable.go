package agredis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

//	type SortedSetCmdable interface {
//		BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) *ZWithKeyCmd
//		BZPopMin(ctx context.Context, timeout time.Duration, keys ...string) *ZWithKeyCmd
//		BZMPop(ctx context.Context, timeout time.Duration, order string, count int64, keys ...string) *ZSliceWithKeyCmd
//		ZAdd(ctx context.Context, key string, members ...Z) *IntCmd
//		ZAddLT(ctx context.Context, key string, members ...Z) *IntCmd
//		ZAddGT(ctx context.Context, key string, members ...Z) *IntCmd
//		ZAddNX(ctx context.Context, key string, members ...Z) *IntCmd
//		ZAddXX(ctx context.Context, key string, members ...Z) *IntCmd
//		ZAddArgs(ctx context.Context, key string, args ZAddArgs) *IntCmd
//		ZAddArgsIncr(ctx context.Context, key string, args ZAddArgs) *FloatCmd
//		ZCard(ctx context.Context, key string) *IntCmd
//		ZCount(ctx context.Context, key, min, max string) *IntCmd
//		ZLexCount(ctx context.Context, key, min, max string) *IntCmd
//		ZIncrBy(ctx context.Context, key string, increment float64, member string) *FloatCmd
//		ZInter(ctx context.Context, store *ZStore) *StringSliceCmd
//		ZInterWithScores(ctx context.Context, store *ZStore) *ZSliceCmd
//		ZInterCard(ctx context.Context, limit int64, keys ...string) *IntCmd
//		ZInterStore(ctx context.Context, destination string, store *ZStore) *IntCmd
//		ZMPop(ctx context.Context, order string, count int64, keys ...string) *ZSliceWithKeyCmd
//		ZMScore(ctx context.Context, key string, members ...string) *FloatSliceCmd
//		ZPopMax(ctx context.Context, key string, count ...int64) *ZSliceCmd
//		ZPopMin(ctx context.Context, key string, count ...int64) *ZSliceCmd
//		ZRange(ctx context.Context, key string, start, stop int64) *StringSliceCmd
//		ZRangeWithScores(ctx context.Context, key string, start, stop int64) *ZSliceCmd
//		ZRangeByScore(ctx context.Context, key string, opt *ZRangeBy) *StringSliceCmd
//		ZRangeByLex(ctx context.Context, key string, opt *ZRangeBy) *StringSliceCmd
//		ZRangeByScoreWithScores(ctx context.Context, key string, opt *ZRangeBy) *ZSliceCmd
//		ZRangeArgs(ctx context.Context, z ZRangeArgs) *StringSliceCmd
//		ZRangeArgsWithScores(ctx context.Context, z ZRangeArgs) *ZSliceCmd
//		ZRangeStore(ctx context.Context, dst string, z ZRangeArgs) *IntCmd
//		ZRank(ctx context.Context, key, member string) *IntCmd
//		ZRankWithScore(ctx context.Context, key, member string) *RankWithScoreCmd
//		ZRem(ctx context.Context, key string, members ...interface{}) *IntCmd
//		ZRemRangeByRank(ctx context.Context, key string, start, stop int64) *IntCmd
//		ZRemRangeByScore(ctx context.Context, key, min, max string) *IntCmd
//		ZRemRangeByLex(ctx context.Context, key, min, max string) *IntCmd
//		ZRevRange(ctx context.Context, key string, start, stop int64) *StringSliceCmd
//		ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) *ZSliceCmd
//		ZRevRangeByScore(ctx context.Context, key string, opt *ZRangeBy) *StringSliceCmd
//		ZRevRangeByLex(ctx context.Context, key string, opt *ZRangeBy) *StringSliceCmd
//		ZRevRangeByScoreWithScores(ctx context.Context, key string, opt *ZRangeBy) *ZSliceCmd
//		ZRevRank(ctx context.Context, key, member string) *IntCmd
//		ZRevRankWithScore(ctx context.Context, key, member string) *RankWithScoreCmd
//		ZScore(ctx context.Context, key, member string) *FloatCmd
//		ZUnionStore(ctx context.Context, dest string, store *ZStore) *IntCmd
//		ZRandMember(ctx context.Context, key string, count int) *StringSliceCmd
//		ZRandMemberWithScores(ctx context.Context, key string, count int) *ZSliceCmd
//		ZUnion(ctx context.Context, store ZStore) *StringSliceCmd
//		ZUnionWithScores(ctx context.Context, store ZStore) *ZSliceCmd
//		ZDiff(ctx context.Context, keys ...string) *StringSliceCmd
//		ZDiffWithScores(ctx context.Context, keys ...string) *ZSliceCmd
//		ZDiffStore(ctx context.Context, destination string, keys ...string) *IntCmd
//		ZScan(ctx context.Context, key string, cursor uint64, match string, count int64) *ScanCmd
//	}

var _ redis.SortedSetCmdable = (*RWClient)(nil)

// ====================== 读操作（使用从节点） ======================
// ZRevRangeByScoreWithScores 查询zset，按照score倒序，从offset偏移，查询count条
func (c *RWClient) ZRevRangeByScoreWithScores(ctx context.Context, key string, args *redis.ZRangeBy) *redis.ZSliceCmd {
	slave := c.getRandomSlave()
	return slave.ZRevRangeByScoreWithScores(ctx, key, args)
}
