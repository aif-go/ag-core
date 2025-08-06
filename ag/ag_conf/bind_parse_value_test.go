package ag_conf

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseValue_Independent(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		target   interface{}
		expected interface{}
		wantErr  bool
	}{
		// 整数测试
		{"int success", "123", new(int), 123, false},
		{"int8 success", "127", new(int8), int8(127), false},
		{"int8 overflow", "128", new(int8), int8(0), true}, // 超出int8范围
		{"int16 success", "32767", new(int16), int16(32767), false},
		{"int16 overflow", "32768", new(int16), int16(0), true}, // 超出int16范围
		{"int32 success", "2147483647", new(int32), int32(2147483647), false},
		{"int32 overflow", "2147483648", new(int32), int32(0), true}, // 超出int32范围
		{"int64 success", "9223372036854775807", new(int64), int64(9223372036854775807), false},
		{"int invalid", "abc", new(int), 0, true},

		// 无符号整数测试
		{"uint success", "123", new(uint), uint(123), false},
		{"uint8 success", "255", new(uint8), uint8(255), false},
		{"uint8 overflow", "256", new(uint8), uint8(0), true}, // 超出uint8范围
		{"uint16 success", "65535", new(uint16), uint16(65535), false},
		{"uint16 overflow", "65536", new(uint16), uint16(0), true}, // 超出uint16范围
		{"uint32 success", "4294967295", new(uint32), uint32(4294967295), false},
		{"uint32 overflow", "4294967296", new(uint32), uint32(0), true}, // 超出uint32范围
		{"uint64 success", "18446744073709551615", new(uint64), uint64(18446744073709551615), false},
		{"uint invalid", "-123", new(uint), uint(0), true},

		// 浮点数测试
		{"float32 success", "3.14", new(float32), float32(3.14), false},
		{"float32 overflow", "1e100", new(float32), float32(0), true}, // 超出float32范围
		{"float64 success", "3.141592653589793", new(float64), 3.141592653589793, false},
		{"float invalid", "3.14a", new(float64), 0.0, true},

		// 布尔值测试
		{"bool true", "true", new(bool), true, false},
		{"bool false", "false", new(bool), false, false},
		{"bool invalid", "yes", new(bool), false, true},

		// 字符串测试
		{"string", "hello", new(string), "hello", false},
		{"empty string", "", new(string), "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.target).Elem()
			param := BindParam{Path: "test.path"}

			err := parseValue(tt.value, v, param)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && !reflect.DeepEqual(v.Interface(), tt.expected) {
				t.Errorf("parseValue() = %v, want %v", v.Interface(), tt.expected)
			}
		})
	}
}

func TestParseValue_UnsupportedType(t *testing.T) {
	var ch chan int
	v := reflect.ValueOf(&ch).Elem()
	param := BindParam{Path: "test.path"}

	err := parseValue("value", v, param)
	if err == nil {
		t.Error("parseValue() should return error for unsupported type")
	} else if !strings.Contains(err.Error(), "unsupported type") {
		t.Errorf("parseValue() should return error containing 'unsupported type', got %v", err)
	}
}
