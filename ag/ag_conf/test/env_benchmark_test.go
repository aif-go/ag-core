package test

import (
	"ag-core/ag/ag_conf"
	"testing"
)

func BenchmarkGetPropertySimple(b *testing.B) {
	env, _ := ag_conf.NewStandardEnvironment()
	ps := &ag_conf.MapPropertySource{
		Source: map[string]any{
			"simple": "value",
		},
	}
	ps.Name = "simple"
	env.GetPropertySources().AddFirst(ps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = env.GetProperty("simple")
	}
}

func BenchmarkGetPropertyPlaceholder(b *testing.B) {
	env, _ := ag_conf.NewStandardEnvironment()
	ps := &ag_conf.MapPropertySource{
		Source: map[string]any{
			"key1": "value1",
			"key2": "${key1}",
		},
	}
	ps.Name = "placeholder"
	env.GetPropertySources().AddFirst(ps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = env.GetProperty("key2")
	}
}

func BenchmarkGetPropertyNestedPlaceholder(b *testing.B) {
	env, _ := ag_conf.NewStandardEnvironment()
	ps := &ag_conf.MapPropertySource{
		Source: map[string]any{
			"a": "A",
			"b": "${a}",
			"c": "${b}",
		},
	}
	ps.Name = "nested"
	env.GetPropertySources().AddFirst(ps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = env.GetProperty("c")
	}
}

func BenchmarkGetPropertyNestedPlaceholder100(b *testing.B) {
	env, _ := ag_conf.NewStandardEnvironment()
	ps := &ag_conf.MapPropertySource{
		Source: map[string]any{
			"a": "A",
			"b": "${a}",
			"c": "${b}",
		},
	}
	ps.Name = "nested"
	env.GetPropertySources().AddFirst(ps)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			_ = env.GetProperty("c")
		}
	}
}

func BenchmarkGetPropertyWithMultipleSources(b *testing.B) {
	env, _ := ag_conf.NewStandardEnvironment()

	// 添加多个PropertySource
	for i := 0; i < 10; i++ {
		ps := &ag_conf.MapPropertySource{
			Source: map[string]any{
				"key": "value" + string(rune('0'+i)),
			},
		}
		ps.Name = "multi_" + string(rune('0'+i))
		env.GetPropertySources().AddFirst(ps)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = env.GetProperty("key")
	}
}

func BenchmarkGetPropertyWithMultipleSources100(b *testing.B) {

	b.ReportAllocs()
	env, _ := ag_conf.NewStandardEnvironment()

	// 添加多个PropertySource
	for i := 0; i < 10; i++ {
		ps := &ag_conf.MapPropertySource{
			Source: map[string]any{
				"key": "value" + string(rune('0'+i)),
			},
		}
		ps.Name = "multi_" + string(rune('0'+i))
		env.GetPropertySources().AddFirst(ps)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 100; j++ {
			_ = env.GetProperty("key")
		}
	}
}
