package test

import (
	"ag-core/contribute/agredis"
	"context"
	"fmt"
	"testing"
	"time"

	// "github.com/go-redis/redis/v8"
	"github.com/redis/go-redis/v9"
)

func TestRedisInit(t *testing.T) {
	var rcli *redis.Client
	rcli = redis.NewClient(nil)
	rcli = redis.NewFailoverClient(nil) // 哨兵模式客户端

	// rcli.Set()
	// rcli.GetDel()
	// rcli.Keys()

	var rclucli *redis.ClusterClient
	rclucli = redis.NewClusterClient(nil)         // 集群模式客户端
	rclucli = redis.NewFailoverClusterClient(nil) // 哨兵模式集群客户端

	var runicli redis.UniversalClient
	runicli = redis.NewUniversalClient(nil) // 通用模式客户端

	// TODO 扩展 读写分离客户端

	var ringCli *redis.Ring
	ringCli = redis.NewRing(nil)

	// NewClient、NewFailoverClient、NewClusterClient、NewFailoverClusterClient 还有别的创建client的函数吗

	fmt.Sprintf("%v, %v, %v, %v", rcli, rclucli, runicli, ringCli)
}
func TestRedis1(t *testing.T) {
	rcli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rcli.Close()

	testCase(rcli, t)
}

func TestRWRedis1(t *testing.T) {
	mcli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	scli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	rwcli := &agredis.RWClient{
		UniversalClient: mcli,
	}
	rwcli.AddSlave("slave1", scli)

	testCase(rwcli, t)
}

func testCase(rcli agredis.AgRedisClient, t *testing.T) {

	ctx := context.Background()
	// Increment the counter
	val, err := rcli.Incr(ctx, "aaa").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("aaa: %d\n", val) // aaa: 1 (then 2, 3, ...)

	rcli.MGet(ctx, "aaa")
	res, err := rcli.MGet(ctx, "aaa").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("aaa: %v\n", res)

	rcli.MSet(ctx, "aaa", 1)

	// ttl
	rcli.Set(ctx, "hzw", "hello", time.Millisecond*500)
	r, err := rcli.Get(ctx, "hzw").Result()
	fmt.Printf("hzw:%s %v\n", r, err)

	time.Sleep(time.Second)

	r, err = rcli.Get(ctx, "hzw").Result()
	fmt.Printf("hzw:%s %v\n", r, err)

	// getdel
	r, _ = rcli.GetDel(ctx, "aaa").Result()
	fmt.Printf("aaa:%s\n", r)

	exis, _ := rcli.Exists(ctx, "aaa").Result()
	fmt.Printf("aaa:%d\n", exis)

	val, err = rcli.Incr(ctx, "bbb").Result()
	if err != nil {
		panic(err)
	}
	fmt.Printf("bbb: %d\n", val)

}
