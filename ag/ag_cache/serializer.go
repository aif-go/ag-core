package ag_cache

import (
	"encoding/json"
	"fmt"
)

// Serializer converts between T and []byte for cache storage.
// Implementations must be thread-safe.
type Serializer[T any] interface {
	Marshal(v T) ([]byte, error)
	Unmarshal(data []byte) (*T, error)
}

// DefaultSerializer uses type-switch fast path for basic types,
// falling back to encoding/json for complex types.
func DefaultSerializer[T any]() Serializer[T] {
	return &defaultSerializer[T]{}
}

type defaultSerializer[T any] struct{}

func (s *defaultSerializer[T]) Marshal(v T) ([]byte, error) {
	// Type-switch fast path: avoid JSON overhead for basic types.
	switch val := any(v).(type) {
	case string:
		return []byte(val), nil
	case []byte:
		cp := make([]byte, len(val))
		copy(cp, val)
		return cp, nil
	case int:
		return []byte(fmt.Sprintf("%d", val)), nil
	case int64:
		return []byte(fmt.Sprintf("%d", val)), nil
	case float64:
		return []byte(fmt.Sprintf("%f", val)), nil
	case bool:
		if val {
			return []byte("1"), nil
		}
		return []byte("0"), nil
	default:
		return json.Marshal(v)
	}
}

func (s *defaultSerializer[T]) Unmarshal(data []byte) (*T, error) {
	var v T
	// Type-switch fast path.
	switch any(&v).(type) {
	case *string:
		str := string(data)
		return any(&str).(*T), nil
	case *[]byte:
		cp := make([]byte, len(data))
		copy(cp, data)
		return any(&cp).(*T), nil
	case *int:
		var n int
		if _, err := fmt.Sscanf(string(data), "%d", &n); err != nil {
			return nil, err
		}
		return any(&n).(*T), nil
	case *int64:
		var n int64
		if _, err := fmt.Sscanf(string(data), "%d", &n); err != nil {
			return nil, err
		}
		return any(&n).(*T), nil
	case *float64:
		var f float64
		if _, err := fmt.Sscanf(string(data), "%f", &f); err != nil {
			return nil, err
		}
		return any(&f).(*T), nil
	case *bool:
		b := string(data) == "1" || string(data) == "true"
		return any(&b).(*T), nil
	default:
		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		return &v, nil
	}
}
