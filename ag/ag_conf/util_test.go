package ag_conf

import (
	"strings"
	"testing"
)

// hasPrefixIgnoreCase功能测试
func TestHasPrefixIgnoreCase(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		prefix string
		want   bool
	}{
		{"空字符串", "", "", true},
		{"相同字符串", "Hello", "Hello", true},
		{"大小写不同", "HELLO", "hello", true},
		{"前缀更长", "Hi", "Hello", false},
		{"前缀匹配但字符串更短", "Hi", "Hello", false},
		{"字符串更短但前缀匹配", "Hello", "He", true},
		{"混合大小写", "hElLo", "HeLlO", true},
		{"非ASCII字符", "中文测试", "中文", true},
		{"非ASCII大小写", "Γειά σου", "Γειά", true},
		{"不匹配前缀", "Hello", "World", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasPrefixIgnoreCase(tt.s, tt.prefix); got != tt.want {
				t.Errorf("hasPrefixIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

// trimPrefixIgnoreCase功能测试
func TestTrimPrefixIgnoreCase(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		prefix string
		want   string
	}{
		{"空前缀", "Hello", "", "Hello"},
		{"相同字符串", "Hello", "Hello", ""},
		{"大小写不同", "HELLO", "hello", ""},
		{"前缀更长", "Hi", "Hello", "Hi"},
		{"部分匹配", "HelloWorld", "Hello", "World"},
		{"混合大小写", "hElLoWorld", "HeLlO", "World"},
		{"非ASCII字符", "中文测试", "中文", "测试"},
		{"非ASCII大小写", "Γειά σου", "Γειά", " σου"},
		{"不匹配前缀", "Hello", "World", "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trimPrefixIgnoreCase(tt.s, tt.prefix); got != tt.want {
				t.Errorf("trimPrefixIgnoreCase(%q, %q) = %q, want %q", tt.s, tt.prefix, got, tt.want)
			}
		})
	}
}

// hasPrefixIgnoreCase性能测试
func BenchmarkHasPrefixIgnoreCase(b *testing.B) {
	testCases := []struct {
		name   string
		s      string
		prefix string
	}{
		{"ASCII短字符串", "Hello", "He"},
		{"ASCII长字符串", "This is a long string for testing", "This is"},
		{"非ASCII字符串", "这是一个测试字符串", "这是"},
		{"混合字符串", "Hello 这是一个测试", "Hello"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				hasPrefixIgnoreCase(tc.s, tc.prefix)
			}
		})
	}
}

// trimPrefixIgnoreCase性能测试
func BenchmarkTrimPrefixIgnoreCase(b *testing.B) {
	testCases := []struct {
		name   string
		s      string
		prefix string
	}{
		{"ASCII短字符串", "Hello", "He"},
		{"ASCII长字符串", "This is a long string for testing", "This is"},
		{"非ASCII字符串", "这是一个测试字符串", "这是"},
		{"混合字符串", "Hello 这是一个测试", "Hello"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				trimPrefixIgnoreCase(tc.s, tc.prefix)
			}
		})
	}
}

// 标准库strings.TrimPrefix性能对比测试
func BenchmarkStdTrimPrefix(b *testing.B) {
	testCases := []struct {
		name   string
		s      string
		prefix string
	}{
		{"ASCII短字符串", "Hello", "He"},
		{"ASCII长字符串", "This is a long string for testing", "This is"},
		{"非ASCII字符串", "这是一个测试字符串", "这是"},
		{"混合字符串", "Hello 这是一个测试", "Hello"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				strings.TrimPrefix(tc.s, tc.prefix)
			}
		})
	}
}

// 标准库strings.HasPrefix性能对比测试
func BenchmarkStdHasPrefix(b *testing.B) {
	testCases := []struct {
		name   string
		s      string
		prefix string
	}{
		{"ASCII短字符串", "Hello", "He"},
		{"ASCII长字符串", "This is a long string for testing", "This is"},
		{"非ASCII字符串", "这是一个测试字符串", "这是"},
		{"混合字符串", "Hello 这是一个测试", "Hello"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				strings.HasPrefix(tc.s, tc.prefix)
			}
		})
	}
}
