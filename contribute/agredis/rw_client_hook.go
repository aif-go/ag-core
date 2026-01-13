package agredis

import (
	"github.com/redis/go-redis/v9"
)

// AddHook
func (c *RWClient) AddHook(hook redis.Hook) {
	c.UniversalClient.AddHook(hook)
	for _, slave := range c.slaves {
		slave.AddHook(hook)
	}
}
