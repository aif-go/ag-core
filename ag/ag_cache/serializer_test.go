package ag_cache_test

import (
	"bytes"
	"testing"

	"github.com/aif-go/ag-core/ag/ag_cache"
)

// TestSerializer_BasicTypes 表驱动覆盖基础类型 fast path（string/[]byte/int/int64/float64/bool）
// 的 Marshal→Unmarshal 往返一致性。
func TestSerializer_BasicTypes(t *testing.T) {
	type testCase struct {
		name string
		val  any
	}
	tests := []testCase{
		{"string", "hello"},
		{"string-empty", ""},
		{"string-unicode", "缓存-中文-emoji🚀"},
		{"bytes", []byte{0x00, 0x01, 0xFE, 0xFF}},
		{"bytes-empty", []byte{}},
		{"int", 42},
		{"int-zero", 0},
		{"int-negative", -7},
		{"int64", int64(1<<40 + 123)},
		{"float64", 3.14159},
		{"float64-zero", 0.0},
		{"bool-true", true},
		{"bool-false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.val.(type) {
			case string:
				roundTrip[string](t, v)
			case []byte:
				roundTrip[[]byte](t, v)
			case int:
				roundTrip[int](t, v)
			case int64:
				roundTrip[int64](t, v)
			case float64:
				roundTrip[float64](t, v)
			case bool:
				roundTrip[bool](t, v)
			default:
				t.Fatalf("unsupported test type %T", tt.val)
			}
		})
	}
}

// roundTrip 断言 Marshal→Unmarshal 后与原值相等。
func roundTrip[T any](t *testing.T, v T) {
	t.Helper()
	s := ag_cache.DefaultSerializer[T]()
	data, err := s.Marshal(v)
	if err != nil {
		t.Fatalf("Marshal(%v): %v", v, err)
	}
	got, err := s.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal(%q): %v", data, err)
	}
	if !equal(v, *got) {
		t.Fatalf("round trip mismatch: want %v, got %v (data=%q)", v, *got, data)
	}
}

// equal 支持基础类型的泛型比较（[]byte 用 bytes.Equal）。
func equal[T any](a, b T) bool {
	switch av := any(a).(type) {
	case []byte:
		return bytes.Equal(av, any(b).([]byte))
	default:
		return any(a) == any(b)
	}
}

// TestSerializer_JSONType 复杂类型走 encoding/json 路径的往返。
func TestSerializer_JSONType(t *testing.T) {
	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	u := User{Name: "Alice", Age: 30}
	s := ag_cache.DefaultSerializer[User]()

	data, err := s.Marshal(u)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	got, err := s.Unmarshal(data)
	if err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if *got != u {
		t.Fatalf("round trip mismatch: want %+v, got %+v", u, *got)
	}
}

// TestSerializer_JSONInvalid 复杂类型 Unmarshal 非法 JSON 应返回错误。
func TestSerializer_JSONInvalid(t *testing.T) {
	s := ag_cache.DefaultSerializer[struct {
		A int `json:"a"`
	}]()
	if _, err := s.Unmarshal([]byte("not-json")); err == nil {
		t.Fatal("Unmarshal invalid JSON should error")
	}
}
