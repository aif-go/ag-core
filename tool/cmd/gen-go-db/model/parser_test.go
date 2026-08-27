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

// TestGetGoType_Decimal 验证 SQL 类型到 Go 类型的映射，重点是 decimal 映射为 decimal.Decimal，
// float/double 保持 float64 不回归。
func TestGetGoType_Decimal(t *testing.T) {
	tests := []struct {
		sqlType string
		want    string
	}{
		{"decimal", "decimal.Decimal"},
		{"float", "float64"},
		{"float32", "float64"},
		{"double", "float64"},
		{"float64", "float64"},
		{"int", "int"},
		{"int64", "int64"},
		{"bigint", "int64"},
		{"varchar", "string"},
		{"datetime", "time.Time"},
	}
	for _, tt := range tests {
		if got := getGoType(tt.sqlType); got != tt.want {
			t.Errorf("getGoType(%q) = %q; want %q", tt.sqlType, got, tt.want)
		}
	}
}

// TestParseYAML_Decimal 验证 decimal 列解析为 decimal.Decimal 且自动导入 shopspring/decimal。
func TestParseYAML_Decimal(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "decimal_type.yaml")

	yamlContent := `table_name: tm_decimal_type
columns:
- name: salary
  type: decimal
  length: 10,2
- name: score
  type: decimal
  length: "8,2"
- name: ratio
  type: decimal
  length: 100
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 返回错误: %v", err)
	}

	if len(data.Columns) != 3 {
		t.Fatalf("列数量 = %d; want 3", len(data.Columns))
	}
	for _, col := range data.Columns {
		if col.GoType != "decimal.Decimal" {
			t.Errorf("列 %s GoType = %q; want %q", col.Name, col.GoType, "decimal.Decimal")
		}
	}

	imports := data.ModelTemplateData.ImportPackages
	if !contains(imports, "github.com/shopspring/decimal") {
		t.Errorf("ImportPackages = %v; 应包含 %q", imports, "github.com/shopspring/decimal")
	}

	// 验证 length 解析：字符串、带精度字符串、纯数字
	wantLengths := []string{"10,2", "8,2", "100"}
	for i, col := range data.Columns {
		if col.Length != wantLengths[i] {
			t.Errorf("列 %s Length = %q; want %q", col.Name, col.Length, wantLengths[i])
		}
	}
}

// TestParseYAML_DecimalPointer 验证 *decimal.Decimal 指针类型也会自动导入 decimal 包。
func TestParseYAML_DecimalPointer(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "decimal_pointer.yaml")

	yamlContent := `table_name: tm_decimal_pointer
columns:
- name: salary
  type: "*decimal.Decimal"
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 返回错误: %v", err)
	}

	if data.Columns[0].GoType != "*decimal.Decimal" {
		t.Errorf("GoType = %q; want %q", data.Columns[0].GoType, "*decimal.Decimal")
	}
	imports := data.ModelTemplateData.ImportPackages
	if !contains(imports, "github.com/shopspring/decimal") {
		t.Errorf("ImportPackages = %v; 应包含 %q", imports, "github.com/shopspring/decimal")
	}
}

// TestGenerateGormTag_TypeTag 验证 gorm tag 长度处理规则（官方标准方案）：
// string → size:<length>；decimal.Decimal（含指针）→ type:decimal(p,s);precision:p;scale:s；
// 其他类型固定大小，不生成长度相关 tag。
func TestGenerateGormTag_TypeTag(t *testing.T) {
	t.Run("string 列带 length 生成 size:length", func(t *testing.T) {
		col := table.ColumnData{Name: "name", Type: "string", GoType: "string", Length: "20"}
		tag := generateGormTag(&col, nil)
		if !strings.Contains(tag, "size:20") {
			t.Errorf("gorm tag 应包含 size:20, got: %s", tag)
		}
		if strings.Contains(tag, "type:") {
			t.Errorf("string 列不应生成 type tag, got: %s", tag)
		}
	})

	t.Run("decimal 列带 length 生成 type:decimal(p,s);precision:p;scale:s", func(t *testing.T) {
		col := table.ColumnData{Name: "salary", Type: "decimal", GoType: "decimal.Decimal", Length: "10,2"}
		tag := generateGormTag(&col, nil)
		for _, want := range []string{"type:decimal(10,2)", "precision:10", "scale:2"} {
			if !strings.Contains(tag, want) {
				t.Errorf("gorm tag 应包含 %q, got: %s", want, tag)
			}
		}
	})

	t.Run("decimal 列 length 无小数位生成 scale:0", func(t *testing.T) {
		col := table.ColumnData{Name: "amount", Type: "decimal", GoType: "decimal.Decimal", Length: "18"}
		tag := generateGormTag(&col, nil)
		for _, want := range []string{"type:decimal(18)", "precision:18", "scale:0"} {
			if !strings.Contains(tag, want) {
				t.Errorf("gorm tag 应包含 %q, got: %s", want, tag)
			}
		}
	})

	t.Run("decimal 列无 length 生成 type:decimal", func(t *testing.T) {
		col := table.ColumnData{Name: "salary", Type: "decimal", GoType: "decimal.Decimal"}
		tag := generateGormTag(&col, nil)
		if !strings.Contains(tag, "type:decimal") {
			t.Errorf("decimal 无 length 也应生成 type:decimal, got: %s", tag)
		}
	})

	t.Run("*decimal.Decimal 指针列生成 type:decimal(p,s);precision:p;scale:s", func(t *testing.T) {
		col := table.ColumnData{Name: "salary", Type: "*decimal.Decimal", GoType: "*decimal.Decimal", Length: "10,2"}
		tag := generateGormTag(&col, nil)
		for _, want := range []string{"type:decimal(10,2)", "precision:10", "scale:2"} {
			if !strings.Contains(tag, want) {
				t.Errorf("gorm tag 应包含 %q, got: %s", want, tag)
			}
		}
	})

	t.Run("其他固定大小类型带 length 不生成 type/size/precision tag", func(t *testing.T) {
		cases := []table.ColumnData{
			{Name: "id", Type: "int64", GoType: "int64", Length: "20"},
			{Name: "active", Type: "bool", GoType: "bool", Length: "1"},
			{Name: "ratio", Type: "float64", GoType: "float64", Length: "10,2"},
			{Name: "created_at", Type: "time", GoType: "time.Time", Length: "6"},
		}
		for _, col := range cases {
			tag := generateGormTag(&col, nil)
			if strings.Contains(tag, "type:") || strings.Contains(tag, "size:") || strings.Contains(tag, "precision:") {
				t.Errorf("列 %s(%s) 不应生成 type/size/precision tag, got: %s", col.Name, col.GoType, tag)
			}
		}
	})
}

// TestSplitDecimalLen 验证 decimal length 的拆分。
func TestSplitDecimalLen(t *testing.T) {
	tests := []struct {
		length string
		p, s   string
	}{
		{"10,2", "10", "2"},
		{"18", "18", ""},
		{" 12, 4 ", "12", "4"},
	}
	for _, tt := range tests {
		p, s := splitDecimalLen(tt.length)
		if p != tt.p || s != tt.s {
			t.Errorf("splitDecimalLen(%q) = (%q,%q); want (%q,%q)", tt.length, p, s, tt.p, tt.s)
		}
	}
}

// TestGetZeroCheck_Decimal 验证 decimal.Decimal 与 *decimal.Decimal 的零值判断形式。
func TestGetZeroCheck_Decimal(t *testing.T) {
	col := table.ColumnData{GoType: "decimal.Decimal", JsonTag: "salary"}
	if got := getZeroCheck("xxx", col); got != "xxx.salary.IsZero()" {
		t.Errorf("decimal.Decimal 应生成 IsZero 判断, got: %s", got)
	}

	ptrCol := table.ColumnData{GoType: "*decimal.Decimal", JsonTag: "salary"}
	if got := getZeroCheck("xxx", ptrCol); got != "xxx.salary == nil" {
		t.Errorf("*decimal.Decimal 应生成 == nil 判断, got: %s", got)
	}
}

// TestNormalizeGoType 验证原始类型词汇到 Go 类型的归一化（* 前缀透传，其余走 getGoType 映射）。
func TestNormalizeGoType(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"decimal", "decimal.Decimal"},
		{"time", "time.Time"},
		{"int64", "int64"},
		{"string", "string"},
		{"*time.Time", "*time.Time"},
		{"*decimal.Decimal", "*decimal.Decimal"},
	}
	for _, tt := range tests {
		if got := normalizeGoType(tt.in); got != tt.want {
			t.Errorf("normalizeGoType(%q) = %q; want %q", tt.in, got, tt.want)
		}
	}
}

// TestParseYAML_WhereParams_NormalizeGoType 验证 Where_params 的参数类型归一化为 Go 类型并补齐 import。
func TestParseYAML_WhereParams_NormalizeGoType(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "dynamic_decimal.yaml")

	yamlContent := `table_name: tm_dynamic
columns:
- name: id
  type: int64
- name: salary
  type: decimal
- name: create_time
  type: time
self_query_rules:
  FindBySalary:
    select_fields: '*'
    page: false
    dynamic_sql: true
    sql_template: "SELECT * FROM tm_dynamic WHERE salary = @Salary"
    Where_params:
    - colname: salary
      paraname: Salary
      slice: false
      type: decimal
    - colname: create_time
      paraname: CreateTime
      slice: false
      type: time
`
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 返回错误: %v", err)
	}

	if len(data.SelfQueries) != 1 {
		t.Fatalf("SelfQueries 数量 = %d; want 1", len(data.SelfQueries))
	}
	q := data.SelfQueries[0]
	if len(q.WhereColFields) != 2 {
		t.Fatalf("WhereColFields 数量 = %d; want 2", len(q.WhereColFields))
	}

	wantTypes := map[string]string{
		"Salary":     "decimal.Decimal",
		"CreateTime": "time.Time",
	}
	for _, wc := range q.WhereColFields {
		if want, ok := wantTypes[wc.FieldName]; ok {
			if wc.GoType != want {
				t.Errorf("参数 %s GoType = %q; want %q", wc.FieldName, wc.GoType, want)
			}
		}
	}

	imports := data.ModelTemplateData.ImportPackages
	for _, want := range []string{"github.com/shopspring/decimal", "time"} {
		if !contains(imports, want) {
			t.Errorf("ImportPackages = %v; 应包含 %q", imports, want)
		}
	}
}
