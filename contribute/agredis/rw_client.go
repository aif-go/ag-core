package agredis

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	// "github.com/go-redis/redis/v8"
	"github.com/redis/go-redis/v9"
)

var (
	_ AgRedisClient = (*RWClient)(nil)
)

// Deprecated RWClient 读写分离Redis客户端 (此RWClient因项目需要设计导致的，不建议使用)
type RWClient struct {
	redis.UniversalClient // 主节点（写操作），所有api默认使用主节点
	// master                redis.UniversalClient            // 主节点（写操作），所有api默认使用主节点
	slaves map[string]redis.UniversalClient // 从节点（读操作），键为节点ID
}

func (c *RWClient) AddSlave(id string, slave redis.UniversalClient) {
	c.slaves[id] = slave
}

// getRandomSlave 获取一个随机的从节点
func (c *RWClient) getRandomSlave() AgRedisClient {
	if len(c.slaves) == 0 {
		return c.UniversalClient
	}
	var keys []string
	for k := range c.slaves {
		keys = append(keys, k)
	}
	return c.slaves[keys[rand.Intn(len(keys))]]
}

// ====================== 管理方法 ======================
// Ping 检查所有节点连接
func (c *RWClient) Ping(ctx context.Context) *redis.StatusCmd {
	mResult := c.UniversalClient.Ping(ctx)
	if mResult.Err() != nil {
		return mResult // 主挂了，直接失败
	}

	// 无从库
	if len(c.slaves) == 0 {
		return mResult
	}

	// 快速检查从库（设置短超时，避免拖慢）
	tctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	slaveOK := false
	for _, slave := range c.slaves {
		sresult := slave.Ping(tctx)
		if _, err := sresult.Result(); err == nil {
			slaveOK = true
			break // 有一个从库活就行
		}
	}

	// 若主活但从全挂，视为整体不可用
	if !slaveOK {
		fakeCmd := &redis.StatusCmd{}
		fakeCmd.SetErr(errors.New("RWClient unhealthy: master is up but all slaves are down"))
		return fakeCmd
	}

	// 正常：返回主库的结果
	return mResult
}

// Echo 此方法用于检查redis服务的状态，没有返回error则认为服务正常
// 如果主库不可用 → 整体不可用（返回主库错误）
// 如果主库可用，但所有从库都不可用 → 整体不可用（因为读写分离已失效）
// 如果主库 + 至少一个从库可用 → 整体可用（返回主库结果）
func (c *RWClient) Echo(ctx context.Context, message interface{}) *redis.StringCmd {
	mResult := c.UniversalClient.Echo(ctx, message)
	if mResult.Err() != nil {
		return mResult // 主挂了，直接失败
	}

	// 无从库
	if len(c.slaves) == 0 {
		return mResult
	}

	// 快速检查从库（设置短超时，避免拖慢）
	echoCtx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	slaveOK := false
	for _, slave := range c.slaves {
		sresult := slave.Echo(echoCtx, message)
		if _, err := sresult.Result(); err == nil {
			slaveOK = true
			break // 有一个从库活就行
		}
	}

	// 若主活但从全挂，视为整体不可用
	if !slaveOK {
		fakeCmd := &redis.StringCmd{}
		fakeCmd.SetErr(errors.New("RWClient unhealthy: master is up but all slaves are down"))
		return fakeCmd
	}

	// 正常：返回主库的 Echo 结果
	return mResult
}

// Close 关闭所有连接
func (c *RWClient) Close() error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(c.slaves)+1)

	// 并发关闭主节点
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := c.UniversalClient.Close(); err != nil {
			errChan <- fmt.Errorf("主节点关闭失败: %w", err)
		}
	}()

	// 并发关闭所有从节点
	for id, slave := range c.slaves {
		wg.Add(1)
		go func(id string, slave redis.UniversalClient) {
			defer wg.Done()
			if err := slave.Close(); err != nil {
				errChan <- fmt.Errorf("从节点 %s 关闭失败: %w", id, err)
			}
		}(id, slave)
	}

	// 等待所有关闭操作完成
	wg.Wait()
	close(errChan)

	// 收集所有错误
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

// PoolStats 获取主节点连接池统计信息 TODO 连接池统计信息怎么统计slave节点，默认只统计主节点
// func (c *RWClient) PoolStats() *redis.PoolStats {
// 	return c.UniversalClient.PoolStats()
// }

// Master 获取主节点客户端（用于直接访问）
func (c *RWClient) Master() AgRedisClient {
	return c.UniversalClient
}

// Slaves 获取从节点客户端列表（用于直接访问）
func (c *RWClient) Slaves() map[string]AgRedisClient {
	result := make(map[string]AgRedisClient, len(c.slaves))
	for id, slave := range c.slaves {
		result[id] = slave
	}
	return result
}
