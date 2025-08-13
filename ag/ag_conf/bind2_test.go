package ag_conf

import (
	"fmt"
	"testing"
)

func TestGetDescendantSubKeysOfName_ArrayCases(t *testing.T) {
	keys := []string{
		// "a.b.c",
		// "a.b.d",
		// "a.e.f",
		"x.y[0].z.w[0][0]",
		"x.y[1].z.w[0][0]",
		"x.y[1].z.w[0][1]",
		"x.y[1].z.w[1][3]",
	}
	var subKeys []string
	do := func(tag string) {
		subKeys = getDescendantSubKeysOfName(tag, keys)
		fmt.Printf("tag=%-15s,subKeys=%v\n", tag, subKeys)
	}

	do("x")
	do("x.y[1]")
	do("x.y[1].z")
	do("x.y[1].z.w[0]")
	do("x.y[1].z.w[1]")
	do("x.y[1].z.w[2]")
	// ======
	// tag=x              ,subKeys=[y]
	// tag=x.y[1]         ,subKeys=[z]
	// tag=x.y[1].z       ,subKeys=[w]
	// tag=x.y[1].z.w[0]  ,subKeys=[[0] [1]]
	// tag=x.y[1].z.w[1]  ,subKeys=[[3]]
	// tag=x.y[1].z.w[2]  ,subKeys=[]
}

func TestGetDescendantSubKeysOfName_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		parent   string
		keys     []string
		expected []string
	}{
		{
			name:     "empty input", // 测试空输入情况
			parent:   "parent",
			keys:     []string{},
			expected: []string{},
		},
		{
			name:     "no matches", // 测试无匹配键的情况
			parent:   "parent",
			keys:     []string{"other", "another.key"},
			expected: []string{},
		},
		{
			name:   "single level matches", // 测试单层级键匹配
			parent: "parent",
			keys: []string{
				"parent.child1",
				"parent.child2",
				"parent.child3",
			},
			expected: []string{"child1", "child2", "child3"},
		},
		{
			name:   "nested matches", // 测试嵌套键匹配
			parent: "parent",
			keys: []string{
				"parent.child1.grandchild",
				"parent.child2.grandchild",
				"parent.child3.grandchild",
			},
			expected: []string{"child1", "child2", "child3"},
		},
		{
			name:   "array index no index case", // 测试一维数组索引
			parent: "parent",
			keys: []string{
				"parent[0].child",
				"parent[1].child",
				"parent[2].child",
			},
			expected: []string{},
		},
		{
			name:   "two-dimensional array no index case", // 测试标准二维数组
			parent: "matrix",                              // 切片parent应该含有索引
			keys: []string{
				"matrix[0][0].id",
				"matrix[0][1].id",
				"matrix[1][0].value",
				"matrix[2][0].name",
			},
			expected: []string{},
		},
		{
			name:   "array index cases", // 测试一维数组索引
			parent: "parent[0]",
			keys: []string{
				"parent[0].child",
				"parent[1].child",
				"parent[2].child",
			},
			expected: []string{"child"},
		},
		{
			name:   "two-dimensional array cases", // 测试标准二维数组
			parent: "matrix[0]",                   // 切片parent应该含有索引
			keys: []string{
				"matrix[0][0].id",
				"matrix[0][1].id",
				"matrix[1][0].value",
				"matrix[2][0].name",
			},
			expected: []string{"[0]", "[1]"},
		},
		{
			name:   "irregular two-dimensional array", // 测试不规则二维数组
			parent: "irregular[1][0]",
			keys: []string{
				"irregular[0][0].a",
				"irregular[0][1].b",
				"irregular[1][0].c",
				"irregular[1][0].d",
				"irregular[2][0].e",
			},
			expected: []string{"c", "d"},
		},
		{
			name:   "mixed cases", // 测试混合键类型
			parent: "parent",
			keys: []string{
				"parent.child1",
				"parent[0].child",
				"parent.child2.grandchild",
				"parent[1].child",
			},
			expected: []string{"child1", "[0]", "child2", "[1]"},
		},
		{
			name:   "case insensitive", // 测试大小写不敏感
			parent: "parent",
			keys: []string{
				"PARENT.child1",
				"parent.CHILD2",
				"Parent.Child3",
			},
			expected: []string{"child1", "CHILD2", "Child3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := getDescendantSubKeysOfName(tt.parent, tt.keys)
			fmt.Printf("name=%-15s\n expect:%v\n actual:%v\n", tt.name, tt.expected, actual)
			if len(actual) != len(tt.expected) {
				t.Errorf("expected %d subkeys, got %d", len(tt.expected), len(actual))
				return
			}
			for i := range actual {
				if actual[i] != tt.expected[i] {
					t.Errorf("at index %d: expected %q, got %q", i, tt.expected[i], actual[i])
				}
			}
		})
	}
}
