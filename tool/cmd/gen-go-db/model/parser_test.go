package model

import (
	"os"
	"path/filepath"
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
