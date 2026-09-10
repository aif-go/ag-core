package ag_cache

import (
	"context"
	"testing"
	"time"
)

// TestEngine_SpiCompile 在编译期断言 Engine SPI 形态：将非 nil 引擎
// 赋给形状精确的本地接口。RED：当 SPI 仍携带旧形态时本文件编译失败
// （Set 带 ttl / 无 TTLSetter / Clear 无 prefix）。
func TestEngine_SpiCompile(t *testing.T) {
	engine := NewMockEngine()

	// Engine.Set 必须取 (ctx, key, value) 无 ttl；Clear 取 prefix。
	var wantEngine interface {
		Get(ctx context.Context, key string) ([]byte, error)
		Set(ctx context.Context, key string, value []byte) error
		Del(ctx context.Context, key string) error
		Clear(ctx context.Context, prefix string) error
		Close() error
	} = engine
	_ = wantEngine

	// TTLSetter 是可选的外部 TTL 能力。
	var wantTTL interface {
		SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
	} = engine
	_ = wantTTL
}

// TestEngine_ClearPrefix 断言 Engine.Clear 接收 namespace prefix。
func TestEngine_ClearPrefix(t *testing.T) {
	engine := NewMockEngine()
	var want interface {
		Clear(ctx context.Context, prefix string) error
	} = engine
	_ = want
}
