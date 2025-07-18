package ag_conf

import (
	"testing"
)

func BenchmarkMatchRefreshKey(b *testing.B) {
	testCases := []struct {
		name string
		s    string
		t    string
	}{
		{"完全匹配", "a.BB.c", "a.BB.c"},
		{"大小写不敏感匹配", "a.BB.c", "a.bB"},
		{"不匹配不完整路径", "a.BB.c", "a.B"},
		{"数组索引完全匹配", "a.b.c[1]", "a.b.c[1]"},
		{"数组索引基础匹配", "a.b.c[1]", "a.b.c"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				matchRefreshKey(tc.s, tc.t)
			}
		})
	}
}

func TestMatchRefreshKey(t *testing.T) {
	tests := []struct {
		name string
		s    string
		t    string
		want bool
	}{
		{
			name: "完全匹配",
			s:    "a.BB.c",
			t:    "a.BB.c",
			want: true,
		},
		{
			name: "大小写不敏感匹配",
			s:    "a.BB.c",
			t:    "a.bB",
			want: true,
		},
		{
			name: "不匹配不完整路径",
			s:    "a.BB.c",
			t:    "a.B",
			want: false,
		},
		{
			name: "数组索引完全匹配",
			s:    "a.b.c[1]",
			t:    "a.b.c[1]",
			want: true,
		},
		{
			name: "数组索引基础匹配",
			s:    "a.b.c[1]",
			t:    "a.b.c",
			want: true,
		},
		{
			name: "不匹配错误路径",
			s:    "a.b.c",
			t:    "x.y.z",
			want: false,
		},
		{
			name: "空字符串",
			s:    "",
			t:    "",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchRefreshKey(tt.s, tt.t); got != tt.want {
				t.Errorf("matchRefreshKey(%q, %q) = %v, want %v", tt.s, tt.t, got, tt.want)
			}
		})
	}
}
