package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// TestParseYAML_PointerTime 验证以 * 前缀开头的 type 列原样透传为 GoType。
func TestParseYAML_PointerTime(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "pointer_type.yaml")

	yamlContent := `table_name: tm_pointer_type
columns:
- name: created_at
  type: "*time.Time"
- name: count
  type: "*int64"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 返回错误: %v", err)
	}

	if len(data.Columns) != 2 {
		t.Fatalf("列数量 = %d; want 2", len(data.Columns))
	}

	timeCol := data.Columns[0]
	if timeCol.GoType != "*time.Time" {
		t.Errorf("GoType = %q; want %q", timeCol.GoType, "*time.Time")
	}
	if timeCol.Type != "*time.Time" {
		t.Errorf("Type = %q; want %q", timeCol.Type, "*time.Time")
	}

	intCol := data.Columns[1]
	if intCol.GoType != "*int64" {
		t.Errorf("GoType = %q; want %q", intCol.GoType, "*int64")
	}
}

// TestParseYAML_TimeRegression 验证非指针类型的固化映射（time→time.Time）未被破坏。
func TestParseYAML_TimeRegression(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "time_regression.yaml")

	yamlContent := `table_name: tm_time_regression
columns:
- name: created_at
  type: time
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 返回错误: %v", err)
	}

	if len(data.Columns) != 1 {
		t.Fatalf("列数量 = %d; want 1", len(data.Columns))
	}

	timeCol := data.Columns[0]
	if timeCol.GoType != "time.Time" {
		t.Errorf("GoType = %q; want %q", timeCol.GoType, "time.Time")
	}
}

// TestParseYAML_Import 验证指针类型列对 time 包的导入逻辑。
func TestParseYAML_Import(t *testing.T) {
	t.Run("指针时间列导入time", func(t *testing.T) {
		tmpDir := t.TempDir()
		yamlPath := filepath.Join(tmpDir, "pointer_time_import.yaml")

		yamlContent := `table_name: tm_pointer_time_import
columns:
- name: created_at
  type: "*time.Time"
`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("写入YAML文件失败: %v", err)
		}

		data, err := ParseYAML(yamlPath, "test-module")
		if err != nil {
			t.Fatalf("ParseYAML 返回错误: %v", err)
		}

		imports := data.ModelTemplateData.ImportPackages
		if !contains(imports, "time") {
			t.Errorf("ImportPackages = %v; 应包含 %q", imports, "time")
		}
	})

	t.Run("非时间指针列不导入time", func(t *testing.T) {
		tmpDir := t.TempDir()
		yamlPath := filepath.Join(tmpDir, "pointer_int64_import.yaml")

		yamlContent := `table_name: tm_pointer_int64_import
columns:
- name: count
  type: "*int64"
`
		if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
			t.Fatalf("写入YAML文件失败: %v", err)
		}

		data, err := ParseYAML(yamlPath, "test-module")
		if err != nil {
			t.Fatalf("ParseYAML 返回错误: %v", err)
		}

		imports := data.ModelTemplateData.ImportPackages
		if contains(imports, "time") {
			t.Errorf("ImportPackages = %v; 不应包含 %q", imports, "time")
		}
	})
}

// TestParseYAML_SelfQueriesOrder 验证 self_query_rules 解析后按 Name 字母序排列，保证生成产物顺序固定。
func TestParseYAML_SelfQueriesOrder(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "self_query_order.yaml")

	yamlContent := `table_name: tm_order
columns:
- name: id
  type: int64
self_query_rules:
  Zebra:
    select_fields: '*'
    page: false
  Apple:
    select_fields: '*'
    page: false
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 返回错误: %v", err)
	}

	if len(data.SelfQueries) != 2 {
		t.Fatalf("SelfQueries 数量 = %d; want 2", len(data.SelfQueries))
	}
	got := []string{data.SelfQueries[0].Name, data.SelfQueries[1].Name}
	want := []string{"Apple", "Zebra"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("SelfQueries[%d].Name = %q; want %q (应按字母序, got=%v)", i, got[i], want[i], got)
		}
	}
}

// TestGenerateGormTag_IndexPriority 验证一列挂多索引时 gorm tag 按索引名排序，保证生成顺序固定。
func TestGenerateGormTag_IndexPriority(t *testing.T) {
	col := table.ColumnData{
		Name: "card_no",
		IndexPriorities: map[string]int{
			"idx_z": 2,
			"idx_a": 1,
		},
	}
	tag := generateGormTag(&col, nil)

	idxAPos := strings.Index(tag, "index:idx_a,priority:1")
	idxZPos := strings.Index(tag, "index:idx_z,priority:2")
	if idxAPos == -1 {
		t.Errorf("tag 应包含 %q, got: %s", "index:idx_a,priority:1", tag)
	}
	if idxZPos == -1 {
		t.Errorf("tag 应包含 %q, got: %s", "index:idx_z,priority:2", tag)
	}
	if idxAPos > idxZPos {
		t.Errorf("tag 中 idx_a 应在 idx_z 之前(按索引名排序), got: %s", tag)
	}
}

// TestGetZeroCheck 验证指针类型列生成 == nil 零值判断，非指针列保持原有判断形式。
func TestGetZeroCheck(t *testing.T) {
	tests := []struct {
		name       string
		goType     string
		jsonTag    string
		want       string
	}{
		{"指针时间列", "*time.Time", "createdAt", "xxx.createdAt == nil"},
		{"时间列", "time.Time", "createdAt", "xxx.createdAt.IsZero()"},
		{"字符串列", "string", "name", `xxx.name == ""`},
		{"布尔列", "bool", "active", "!xxx.active"},
		{"数值列", "int64", "age", "xxx.age == 0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := table.ColumnData{GoType: tt.goType, JsonTag: tt.jsonTag}
			got := getZeroCheck("xxx", col)
			if got != tt.want {
				t.Errorf("getZeroCheck(xxx, {GoType:%q, JsonTag:%q}) = %q; want %q",
					tt.goType, tt.jsonTag, got, tt.want)
			}
		})
	}
}
