package excel

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestCase 测试用例结构
type TestCase struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Input       string      `yaml:"input"`
	Output      *WhereClause `yaml:"output"`
}

// TestParseWhereCondition 测试 ParseWhereCondition 方法
func TestParseWhereCondition(t *testing.T) {

	testCases:=getTestCases()
	// 执行测试并生成YAML文件
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
		    t.Log("测试方法:",tc.Name,"输入条件:",tc.Input)
			result := ParseWhereCondition(tc.Input)

			// t.Log(result)
			
			// 验证结果
			if !compareWhereClause(result, tc.Output) {
				t.Errorf("测试失败: %s\n输入: %s\n期望: %+v\n实际: %+v", tc.Name, tc.Input, tc.Output, result)
			}

			// 生成YAML文件
			// generateYAML(tc, result)
		})
	}
}

// compareWhereClause 比较两个WhereClause是否相等
func compareWhereClause(a, b *WhereClause) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Operator != b.Operator {
		return false
	}
	if len(a.Conditions) != len(b.Conditions) {
		return false
	}
	for i := range a.Conditions {
		if !compareCondition(a.Conditions[i], b.Conditions[i]) {
			return false
		}
	}
	return true
}

// compareCondition 比较两个Condition是否相等
func compareCondition(a, b *Condition) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if a.Operator != b.Operator {
		return false
	}
	if a.Expr != b.Expr {
		return false
	}
	// 比较Field字段
	if a.Field != b.Field {
		return false
	}
	// 比较Values字段
	if len(a.Values) != len(b.Values) {
		return false
	}
	for i := range a.Values {
		if a.Values[i] != b.Values[i] {
			return false
		}
	}
	if len(a.Conditions) != len(b.Conditions) {
		return false
	}
	for i := range a.Conditions {
		if !compareCondition(a.Conditions[i], b.Conditions[i]) {
			return false
		}
	}
	return true
}

// generateYAML 生成YAML文件
func generateYAML(tc TestCase, result *WhereClause) {
	// 创建test目录
	testDir := "test"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		fmt.Printf("创建目录失败: %v\n", err)
		return
	}

	// 生成文件名
	// safeName := sanitizeFilename(tc.Name)
	safeName := tc.Name
	filename := filepath.Join(testDir, safeName+".yaml")

	// 准备YAML内容
	yamlData := map[string]interface{}{
		"name":        tc.Name,
		"description": tc.Description,
		"input":       tc.Input,
		"output":      result,
	}

	// 转换为YAML
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		fmt.Printf("YAML转换失败: %v\n", err)
		return
	}

	// 写入文件
	if err := os.WriteFile(filename, yamlBytes, 0644); err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}

	fmt.Printf("生成测试文件: %s\n", filename)
}

// sanitizeFilename 清理文件名，移除不安全的字符
// func sanitizeFilename(name string) string {
// 	// 简单的文件名清理
// 	safe := make([]rune, 0, len(name))
// 	for _, r := range name {
// 		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
// 			safe = append(safe, r)
// 		} else {
// 			safe = append(safe, '_')
// 		}
// 	}
// 	return string(safe)
// }

// TestGenerateAllYAML 生成所有测试用例的YAML文件
func TestGenerateAllYAML(t *testing.T) {
	// 定义所有测试用例
	testCases := []struct {
		name        string
		description string
		input       string
	}{
		// 1. 普通where条件
		{"普通where条件_单个条件", "测试单个简单的where条件", "id = 1"},
		{"普通where条件_多个AND条件", "测试多个AND连接的where条件", "id = 1 AND name = 'test' AND age > 18"},
		{"普通where条件_多个OR条件", "测试多个OR连接的where条件", "id = 1 OR id = 2 OR id = 3"},
		{"普通where条件_混合AND和OR", "测试AND和OR混合的where条件", "id = 1 AND name = 'test' OR id = 2"},

		// 2. 嵌套条件
		{"嵌套条件_简单括号嵌套", "测试简单的括号嵌套条件", "(id = 1 OR id = 2) AND name = 'test'"},
		{"嵌套条件_多层括号嵌套", "测试多层括号嵌套条件", "((id = 1 OR id = 2) AND name = 'test') OR (age > 18 AND status = 'active')"},
		{"嵌套条件_复杂嵌套", "测试复杂的嵌套条件", "(id = 1 AND (name = 'test' OR age > 18)) OR (status = 'active' AND (type = 'A' OR type = 'B'))"},

		// 3. 包含between and
		{"BetweenAnd_简单between", "测试简单的BETWEEN AND条件", "age BETWEEN 18 AND 60"},
		{"BetweenAnd_多个between", "测试多个BETWEEN AND条件", "age BETWEEN 18 AND 60 AND score BETWEEN 60 AND 100"},
		{"BetweenAnd_嵌套between", "测试嵌套的BETWEEN AND条件", "(age BETWEEN 18 AND 60 OR age BETWEEN 60 AND 80) AND status = 'active'"},

		// 4. 包含 >= <= 等操作
		{"比较操作_大于等于", "测试大于等于操作符", "age >= 18"},
		{"比较操作_小于等于", "测试小于等于操作符", "age <= 60"},
		{"比较操作_多个比较操作", "测试多个比较操作符组合", "age >= 18 AND age <= 60 AND score > 60 AND score < 100"},
		{"比较操作_不等于", "测试不等于操作符", "status != 'deleted' AND status <> 'inactive'"},

		// 5. 包含 in 或者 not in 操作
		{"IN操作_简单IN", "测试简单的IN操作", "id IN (1, 2, 3)"},
		{"IN操作_字符串IN", "测试字符串类型的IN操作", "status IN ('active', 'pending', 'completed')"},
		{"NOTIN操作_简单NOTIN", "测试简单的NOT IN操作", "id NOT IN (1, 2, 3)"},
		{"IN操作_多个IN组合", "测试多个IN操作组合", "id IN (1, 2, 3) AND status IN ('active', 'pending')"},

		// 6. 包含子查询的查询
		{"子查询_简单子查询", "测试简单的子查询", "id IN (SELECT id FROM users WHERE status = 'active')"},
		{"子查询_复杂子查询", "测试复杂的子查询", "id IN (SELECT id FROM orders WHERE amount > 1000 AND created_at > '2024-01-01')"},
		{"子查询_多个子查询", "测试多个子查询组合", "id IN (SELECT id FROM users WHERE status = 'active') AND order_id IN (SELECT id FROM orders WHERE status = 'completed')"},

		// 7. 包含 exists 或者 not exists 操作
		{"EXISTS_简单EXISTS", "测试简单的EXISTS操作", "EXISTS (SELECT 1 FROM orders WHERE user_id = id AND status = 'completed')"},
		{"NOTEXISTS_简单NOTEXISTS", "测试简单的NOT EXISTS操作", "NOT EXISTS (SELECT 1 FROM orders WHERE user_id = id AND status = 'pending')"},
		{"EXISTS_多个EXISTS组合", "测试多个EXISTS操作组合", "EXISTS (SELECT 1 FROM orders WHERE user_id = id) AND NOT EXISTS (SELECT 1 FROM refunds WHERE order_id = id)"},

		// 8. 包含join操作
		{"JOIN_简单JOIN条件", "测试简单的JOIN条件", "users.id = orders.user_id AND users.status = 'active'"},
		{"JOIN_多表JOIN条件", "测试多表JOIN条件", "users.id = orders.user_id AND orders.id = order_items.order_id AND users.status = 'active'"},
		{"JOIN_复杂JOIN条件", "测试复杂的JOIN条件", "(users.id = orders.user_id AND orders.status = 'completed') OR (users.id = refunds.user_id AND refunds.status = 'approved')"},

		// 综合测试
		{"综合测试_复杂综合条件", "测试综合的复杂条件", "(id IN (1, 2, 3) OR id IN (SELECT id FROM users WHERE status = 'active')) AND (age BETWEEN 18 AND 60 OR age >= 65) AND (status = 'active' OR status = 'pending')"},
	}

	// 创建test目录
	testDir := "test"
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("创建目录失败: %v", err)
	}

	// 为每个测试用例生成YAML文件
	for _, tc := range testCases {
		result := ParseWhereCondition(tc.input)

		yamlData := map[string]interface{}{
			"name":        tc.name,
			"description": tc.description,
			"input":       tc.input,
			"output":      result,
		}

		yamlBytes, err := yaml.Marshal(yamlData)
		if err != nil {
			t.Errorf("YAML转换失败: %v", err)
			continue
		}

		// safeName := sanitizeFilename(tc.name)
		safeNme := tc.name
		filename := filepath.Join(testDir, safeName+".yaml")

		if err := os.WriteFile(filename, yamlBytes, 0644); err != nil {
			t.Errorf("写入文件失败: %v", err)
			continue
		}

		t.Logf("生成测试文件: %s", filename)
	}
}


func getTestCases() []TestCase{
		return []TestCase{
		// 1. 普通where条件
		// {
		// 	Name:        "普通where条件 - 单个条件",
		// 	Description: "测试单个简单的where条件",
		// 	Input:       "id = 1",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id = 1",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "普通where条件 - 多个AND条件",
		// 	Description: "测试多个AND连接的where条件",
		// 	Input:       "id = 1 AND name = 'test' AND age > 18",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id = 1",
		// 			},
		// 			{
		// 				Expr: "name = 'test'",
		// 			},
		// 			{
		// 				Expr: "age > 18",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "普通where条件 - 多个OR条件",
		// 	Description: "测试多个OR连接的where条件",
		// 	Input:       "id = 1 OR id = 2 OR id = 3",
		// 	Output: &WhereClause{
		// 		Operator: "OR",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id = 1",
		// 			},
		// 			{
		// 				Expr: "id = 2",
		// 			},
		// 			{
		// 				Expr: "id = 3",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "普通where条件 - 混合AND和OR",
		// 	Description: "测试AND和OR混合的where条件",
		// 	Input:       "id = 1 AND name = 'test' OR id = 2",
		// 	Output: &WhereClause{
		// 		Operator: "OR",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id = 1",
		// 			},
		// 			{
		// 				Expr: "name = 'test'",
		// 			},
		// 			{
		// 				Expr: "id = 2",
		// 			},
		// 		},
		// 	},
		// },

		// // 2. 嵌套条件
		// {
		// 	Name:        "嵌套条件 - 简单括号嵌套",
		// 	Description: "测试简单的括号嵌套条件",
		// 	Input:       "(id = 1 OR id = 2) AND name = 'test'",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Operator: "OR",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "id = 1",
		// 					},
		// 					{
		// 						Expr: "id = 2",
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Expr: "name = 'test'",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "嵌套条件 - 多层括号嵌套",
		// 	Description: "测试多层括号嵌套条件",
		// 	Input:       "((id = 1 OR id = 2) AND name = 'test') OR (age > 18 AND status = 'active')",
		// 	Output: &WhereClause{
		// 		Operator: "OR",
		// 		Conditions: []*Condition{
		// 			{
		// 				Operator: "AND",
		// 				Conditions: []*Condition{
		// 					{
		// 						Operator: "OR",
		// 						Conditions: []*Condition{
		// 							{
		// 								Expr: "id = 1",
		// 							},
		// 							{
		// 								Expr: "id = 2",
		// 							},
		// 						},
		// 					},
		// 					{
		// 						Expr: "name = 'test'",
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Operator: "AND",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "age > 18",
		// 					},
		// 					{
		// 						Expr: "status = 'active'",
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "嵌套条件 - 复杂嵌套",
		// 	Description: "测试复杂的嵌套条件",
		// 	Input:       "(id = 1 AND (name = 'test' OR age > 18)) OR (status = 'active' AND (type = 'A' OR type = 'B'))",
		// 	Output: &WhereClause{
		// 		Operator: "OR",
		// 		Conditions: []*Condition{
		// 			{
		// 				Operator: "AND",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "id = 1",
		// 					},
		// 					{
		// 						Operator: "OR",
		// 						Conditions: []*Condition{
		// 							{
		// 								Expr: "name = 'test'",
		// 							},
		// 							{
		// 								Expr: "age > 18",
		// 							},
		// 						},
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Operator: "AND",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "status = 'active'",
		// 					},
		// 					{
		// 						Operator: "OR",
		// 						Conditions: []*Condition{
		// 							{
		// 								Expr: "type = 'A'",
		// 							},
		// 							{
		// 								Expr: "type = 'B'",
		// 							},
		// 						},
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },

		// // 3. 包含between and
		// {
		// 	Name:        "Between And - 简单between",
		// 	Description: "测试简单的BETWEEN AND条件",
		// 	Input:       "age BETWEEN 18 AND 60",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "age BETWEEN 18 AND 60",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "Between And - 多个between",
		// 	Description: "测试多个BETWEEN AND条件",
		// 	Input:       "age BETWEEN 18 AND 60 AND score BETWEEN 60 AND 100",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "age BETWEEN 18 AND 60",
		// 			},
		// 			{
		// 				Expr: "score BETWEEN 60 AND 100",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "Between And - 嵌套between",
		// 	Description: "测试嵌套的BETWEEN AND条件",
		// 	Input:       "(age BETWEEN 18 AND 60 OR age BETWEEN 60 AND 80) AND status = 'active'",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Operator: "OR",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "age BETWEEN 18 AND 60",
		// 					},
		// 					{
		// 						Expr: "age BETWEEN 60 AND 80",
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Expr: "status = 'active'",
		// 			},
		// 		},
		// 	},
		// },

		// // 4. 包含 >= <= 等操作
		// {
		// 	Name:        "比较操作 - 大于等于",
		// 	Description: "测试大于等于操作符",
		// 	Input:       "age >= 18",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "age >= 18",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "比较操作 - 小于等于",
		// 	Description: "测试小于等于操作符",
		// 	Input:       "age <= 60",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "age <= 60",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "比较操作 - 多个比较操作",
		// 	Description: "测试多个比较操作符组合",
		// 	Input:       "age >= 18 AND age <= 60 AND score > 60 AND score < 100",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "age >= 18",
		// 			},
		// 			{
		// 				Expr: "age <= 60",
		// 			},
		// 			{
		// 				Expr: "score > 60",
		// 			},
		// 			{
		// 				Expr: "score < 100",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "比较操作 - 不等于",
		// 	Description: "测试不等于操作符",
		// 	Input:       "status != 'deleted' AND status <> 'inactive'",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "status != 'deleted'",
		// 			},
		// 			{
		// 				Expr: "status <> 'inactive'",
		// 			},
		// 		},
		// 	},
		// },

		// // 5. 包含 in 或者 not in 操作
		// {
		// 	Name:        "IN操作 - 简单IN",
		// 	Description: "测试简单的IN操作",
		// 	Input:       "id IN (1, 2, 3)",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id IN (1, 2, 3)",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "IN操作 - 字符串IN",
		// 	Description: "测试字符串类型的IN操作",
		// 	Input:       "status IN ('成功', '未明', '失败')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "status IN ('成功', '未明', '失败')",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "NOT IN操作 - 简单NOT IN",
		// 	Description: "测试简单的NOT IN操作",
		// 	Input:       "id NOT IN (1, 2, 3)",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id NOT IN (1, 2, 3)",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "IN操作 - 多个IN组合",
		// 	Description: "测试多个IN操作组合",
		// 	Input:       "id IN (1, 2, 3) AND status IN ('active', 'pending')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id IN (1, 2, 3)",
		// 			},
		// 			{
		// 				Expr: "status IN ('active', 'pending')",
		// 			},
		// 		},
		// 	},
		// },

		// 6. 包含子查询的查询
		// {
		// 	Name:        "子查询 - 简单子查询",
		// 	Description: "测试简单的子查询",
		// 	Input:       "id IN (SELECT id FROM users WHERE status = 'active')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id IN (SELECT id FROM users WHERE status = 'active')",
		// 			},
		// 		},
		// 	},
		// },
		{
			Name:        "子查询 - 复杂子查询",
			Description: "测试复杂的子查询",
			Input:       "id IN (SELECT id FROM orders WHERE amount > 1000 AND created_at > '2024-01-01')",
			Output: &WhereClause{
				Operator: "AND",
				Conditions: []*Condition{
					{
						Expr: "id IN (SELECT id FROM orders WHERE amount > 1000 AND created_at > '2024-01-01')",
					},
				},
			},
		},
		// {
		// 	Name:        "子查询 - 多个子查询",
		// 	Description: "测试多个子查询组合",
		// 	Input:       "id IN (SELECT id FROM users WHERE status = 'active') AND order_id IN (SELECT id FROM orders WHERE status = 'completed')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "id IN (SELECT id FROM users WHERE status = 'active')",
		// 			},
		// 			{
		// 				Expr: "order_id IN (SELECT id FROM orders WHERE status = 'completed')",
		// 			},
		// 		},
		// 	},
		// },

		// // 7. 包含 exists 或者 not exists 操作
		// {
		// 	Name:        "EXISTS - 简单EXISTS",
		// 	Description: "测试简单的EXISTS操作",
		// 	Input:       "EXISTS (SELECT 1 FROM orders WHERE user_id = id AND status = 'completed')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "EXISTS (SELECT 1 FROM orders WHERE user_id = id AND status = 'completed')",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "NOT EXISTS - 简单NOT EXISTS",
		// 	Description: "测试简单的NOT EXISTS操作",
		// 	Input:       "NOT EXISTS (SELECT 1 FROM orders WHERE user_id = id AND status = 'pending')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "NOT EXISTS (SELECT 1 FROM orders WHERE user_id = id AND status = 'pending')",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "EXISTS - 多个EXISTS组合",
		// 	Description: "测试多个EXISTS操作组合",
		// 	Input:       "EXISTS (SELECT 1 FROM orders WHERE user_id = id) AND NOT EXISTS (SELECT 1 FROM refunds WHERE order_id = id)",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "EXISTS (SELECT 1 FROM orders WHERE user_id = id)",
		// 			},
		// 			{
		// 				Expr: "NOT EXISTS (SELECT 1 FROM refunds WHERE order_id = id)",
		// 			},
		// 		},
		// 	},
		// },

		// // 8. 包含join操作
		// {
		// 	Name:        "JOIN - 简单JOIN条件",
		// 	Description: "测试简单的JOIN条件",
		// 	Input:       "users.id = orders.user_id AND users.status = 'active'",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "users.id = orders.user_id",
		// 			},
		// 			{
		// 				Expr: "users.status = 'active'",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "JOIN - 多表JOIN条件",
		// 	Description: "测试多表JOIN条件",
		// 	Input:       "users.id = orders.user_id AND orders.id = order_items.order_id AND users.status = 'active'",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Expr: "users.id = orders.user_id",
		// 			},
		// 			{
		// 				Expr: "orders.id = order_items.order_id",
		// 			},
		// 			{
		// 				Expr: "users.status = 'active'",
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "JOIN - 复杂JOIN条件",
		// 	Description: "测试复杂的JOIN条件",
		// 	Input:       "(users.id = orders.user_id AND orders.status = 'completed') OR (users.id = refunds.user_id AND refunds.status = 'approved')",
		// 	Output: &WhereClause{
		// 		Operator: "OR",
		// 		Conditions: []*Condition{
		// 			{
		// 				Operator: "AND",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "users.id = orders.user_id",
		// 					},
		// 					{
		// 						Expr: "orders.status = 'completed'",
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Operator: "AND",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "users.id = refunds.user_id",
		// 					},
		// 					{
		// 						Expr: "refunds.status = 'approved'",
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
		// {
		// 	Name:        "综合测试 - 复杂综合条件",
		// 	Description: "测试综合的复杂条件",
		// 	Input:       "(id IN (1, 2, 3) OR id IN (SELECT id FROM users WHERE status = 'active')) AND (age BETWEEN 18 AND 60 OR age >= 65) AND (status = 'active' OR status = 'pending')",
		// 	Output: &WhereClause{
		// 		Operator: "AND",
		// 		Conditions: []*Condition{
		// 			{
		// 				Operator: "OR",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "id IN (1, 2, 3)",
		// 					},
		// 					{
		// 						Expr: "id IN (SELECT id FROM users WHERE status = 'active')",
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Operator: "OR",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "age BETWEEN 18 AND 60",
		// 					},
		// 					{
		// 						Expr: "age >= 65",
		// 					},
		// 				},
		// 			},
		// 			{
		// 				Operator: "OR",
		// 				Conditions: []*Condition{
		// 					{
		// 						Expr: "status = 'active'",
		// 					},
		// 					{
		// 						Expr: "status = 'pending'",
		// 					},
		// 				},
		// 			},
		// 		},
		// 	},
		// },
	}
}