package agredis

import (
	"fmt"
	"log/slog"

	"github.com/redis/go-redis/v9"
)

type AgRedisClientBuilder struct {
	Config *AgRedisProperties
}

func (b *AgRedisClientBuilder) Build() (AgRedisClient, error) {
	conf := b.Config

	ctype := conf.Type
	switch ctype {
	// case TypeFailover:
	// case TypeFailoverCluster:
	// case TypeCluster:
	// case TypeSingle:
	case TypeUniversal:
		slog.Info("agredis build universal client")
		return b.buildUniversalClient()
	case TypeRW:
		slog.Info("agredis build rw client")
		return b.buildRWClient()
	default:
		return nil, fmt.Errorf("unknown redis client type: %s", ctype)
	}
}

func (b *AgRedisClientBuilder) buildRWClient() (AgRedisClient, error) {
	mconf := b.Config.Config
	rconfs := b.Config.Replicas

	rwcli := &RWClient{
		slaves: make(map[string]redis.UniversalClient),
	}

	// 主client
	mcli, err := b.buildUniversalCli(mconf)
	if err != nil {
		return nil, err
	}
	rwcli.UniversalClient = mcli

	// 从client
	// add replicas
	for i, rconf := range rconfs {
		scli, err := b.buildUniversalCli(rconf)
		if err != nil {
			return nil, err
		}
		rwcli.AddSlave(fmt.Sprintf("slave%d", i), scli)
	}

	return rwcli, nil
}

func (b *AgRedisClientBuilder) buildUniversalClient() (AgRedisClient, error) {
	conf := b.Config
	mconf := conf.Config

	return b.buildUniversalCli(mconf)
}

func (b *AgRedisClientBuilder) buildUniversalCli(conf AgUniversalOptionsProperties) (cli redis.UniversalClient, rerr error) {
	defer func() {
		// rerr = recover().(error)
		if rec := recover(); rec != nil {
			// 关闭已创建的客户端
			if cli != nil {
				cli.Close()
			}

			// 将 panic 转换为错误
			if err, ok := rec.(error); ok {
				rerr = err
			} else {
				rerr = fmt.Errorf("panic occurred: %v", rec)
			}
		}
	}()

	options := NewUniversalOptionsWithAgUniversalProperties(conf)
	b.enhanceOptions(options) // 增强
	cli = redis.NewUniversalClient(options)
	return
}

// enhanceOptions 增强options
func (b *AgRedisClientBuilder) enhanceOptions(options *redis.UniversalOptions) {
	// 增强options TODO 有需求时再增加，如TLS等配置
}

// CreateClientByBuilder 通过构建器创建AgRedisClient
func CreateClientByBuilder(builder *AgRedisClientBuilder) (AgRedisClient, error) {
	return builder.Build()
}
