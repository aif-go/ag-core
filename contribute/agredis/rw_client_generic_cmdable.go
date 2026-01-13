package agredis

import (
	"context"

	// "github.com/go-redis/redis/v8"
	"github.com/redis/go-redis/v9"
)

func (c *RWClient) Keys(ctx context.Context, pattern string) *redis.StringSliceCmd {
	slave := c.getRandomSlave()
	return slave.Keys(ctx, pattern)
}
