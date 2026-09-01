package ag_cache

import (
	"context"
	"testing"
	"time"
)

// TestEngine_SpiCompile asserts the Engine SPI shape at compile time by
// assigning a non-nil engine to a local interface of the exact desired shape.
// RED: this file fails to compile while the SPI still carries the old shapes
// (Set with ttl / no TTLSetter / no prefix on Clear).
func TestEngine_SpiCompile(t *testing.T) {
	engine := NewMockEngine()

	// Engine.Set must take (ctx, key, value) with no ttl; Clear takes a prefix.
	var wantEngine interface {
		Get(ctx context.Context, key string) ([]byte, error)
		Set(ctx context.Context, key string, value []byte) error
		Del(ctx context.Context, key string) error
		Clear(ctx context.Context, prefix string) error
		Close() error
	} = engine
	_ = wantEngine

	// TTLSetter is the optional external-TTL capability.
	var wantTTL interface {
		SetWithTTL(ctx context.Context, key string, value []byte, ttl time.Duration) error
	} = engine
	_ = wantTTL
}

// TestEngine_ClearPrefix asserts Engine.Clear takes the namespace prefix.
func TestEngine_ClearPrefix(t *testing.T) {
	engine := NewMockEngine()
	var want interface {
		Clear(ctx context.Context, prefix string) error
	} = engine
	_ = want
}
