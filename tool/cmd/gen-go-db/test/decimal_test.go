package test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/model"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

const testYAML = `table_name: payment
columns:
  - name: id
    type: decimal(18,0)
  - name: batch_id
    type: DECIMAL(18,0)
  - name: amount
    type: decimal(18,2)
  - name: fee
    type: decimal(18,2)
  - name: name
    type: varchar(64)
  - name: created_at
    type: datetime
primary_key:
  - id
  - batch_id
indexes:
  - name: idx_amount
    columns:
      - amount
      - fee
self_query_rules:
  FindByAmount:
    select_fields: "*"
    where:
      operator: AND
      conditions:
        - expr: "amount = @Amount"
`

func parseYAMLFromString(t *testing.T, yamlContent string) *table.TableData {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.yaml")
	if err := os.WriteFile(path, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("write yaml file failed: %v", err)
	}
	td, err := model.ParseYAML(path, "")
	if err != nil {
		t.Fatalf("ParseYAML failed: %v", err)
	}
	return td
}

func checkGoType(t *testing.T, td *table.TableData, colName, want string) {
	t.Helper()
	for _, col := range td.Columns {
		if col.Name == colName {
			if col.GoType != want {
				t.Errorf("column %s: GoType = %q, want %q", colName, col.GoType, want)
			}
			return
		}
	}
	t.Errorf("column %s not found", colName)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected %q NOT to contain %q (but it did)", s, substr)
	}
}

func TestGetGoTypeConvergence(t *testing.T) {
	tests := []struct {
		name    string
		sqlType string
		want    string
	}{
		{name: "decimal plain", sqlType: "decimal", want: "decimal.Decimal"},
		{name: "decimal with precision", sqlType: "decimal(18,2)", want: "decimal.Decimal"},
		{name: "decimal uppercase", sqlType: "DECIMAL", want: "decimal.Decimal"},
		{name: "decimal uppercase precision", sqlType: "DECIMAL(18,2)", want: "decimal.Decimal"},
		{name: "float unchanged", sqlType: "float", want: "float64"},
		{name: "int unchanged", sqlType: "int", want: "int"},
		{name: "varchar unchanged", sqlType: "varchar", want: "string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := "table_name: test\ncolumns:\n  - name: col\n    type: " + tt.sqlType + "\n"
			td := parseYAMLFromString(t, yaml)
			if len(td.Columns) != 1 {
				t.Fatalf("expected 1 column, got %d", len(td.Columns))
			}
			if td.Columns[0].GoType != tt.want {
				t.Errorf("getGoType(%q) via ParseYAML = %q, want %q", tt.sqlType, td.Columns[0].GoType, tt.want)
			}
		})
	}
}

func TestDecimalFullGeneration(t *testing.T) {
	td := parseYAMLFromString(t, testYAML)

	if len(td.Columns) != 6 {
		t.Fatalf("expected 6 columns, got %d", len(td.Columns))
	}

	// 验证类型映射收敛
	checkGoType(t, td, "id", "decimal.Decimal")
	checkGoType(t, td, "batch_id", "decimal.Decimal")
	checkGoType(t, td, "amount", "decimal.Decimal")
	checkGoType(t, td, "fee", "decimal.Decimal")
	checkGoType(t, td, "name", "string")
	checkGoType(t, td, "created_at", "time.Time")

	// 验证 import 注入
	if td.ModelTemplateData == nil {
		t.Fatal("ModelTemplateData is nil")
	}
	found := false
	for _, pkg := range td.ModelTemplateData.ImportPackages {
		if pkg == "github.com/shopspring/decimal" {
			found = true
			break
		}
	}
	if !found {
		t.Error("ModelTemplateData.ImportPackages missing github.com/shopspring/decimal")
	}

	// 生成 Model 代码
	modelCode := model.GetModelTemplate(td)
	if modelCode == "" {
		t.Fatal("GetModelTemplate returned empty")
	}
	t.Logf("Model code length: %d bytes", len(modelCode))

	// 验证 model 中 decimal 字段使用 IsZero 而不是 != 0
	assertNotContains(t, modelCode, "payment.Amount == 0")
	assertContains(t, modelCode, ".IsZero()")

	// 生成 DAO 代码
	daoCode := dao.GetDaoTemplate(td)
	if daoCode == "" {
		t.Fatal("GetDaoTemplate returned empty")
	}
	t.Logf("DAO code length: %d bytes", len(daoCode))

	// 验证所有 6 处 switch 都使用 IsZero
	assertContains(t, daoCode, ".IsZero()")
	assertNotContains(t, daoCode, "payment.Amount == 0")
}

func TestDecimalZeroValueCheck(t *testing.T) {
	td := parseYAMLFromString(t, testYAML)
	daoCode := dao.GetDaoTemplate(td)

	// daoT:22 generateZeroValueCheck → 多列 OR 判断
	assertContains(t, daoCode, "Amount.IsZero()")
	assertContains(t, daoCode, "Fee.IsZero()")

	// daoT:164 第一主键 id (decimal(18,0))
	assertContains(t, daoCode, "Id.IsZero()")

	// daoT:187 第二+主键 batch_id (DECIMAL(18,0))
	assertContains(t, daoCode, "BatchId.IsZero()")

	// daoT:234 索引第一列 amount
	assertContains(t, daoCode, "Amount.IsZero()")

	// daoT:255 索引第二+列 fee
	assertContains(t, daoCode, "Fee.IsZero()")

	// 确保 varchar 和 datetime 列不受影响
	assertContains(t, daoCode, "Name != \"\"")
	assertNotContains(t, daoCode, "Name == 0")
}

func TestDecimalGeneratedCodeCompiles(t *testing.T) {
	td := parseYAMLFromString(t, testYAML)

	modelCode := model.GetModelTemplate(td)

	dir := t.TempDir()

	goModContent := `module testdecimal

go 1.24.8

require github.com/shopspring/decimal v1.4.0
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(goModContent), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "payment_model.go"), []byte(modelCode), 0644); err != nil {
		t.Fatal(err)
	}

	// 为 package model 创建独立的验证文件
	checkCode := `package model

func CheckDecimalCompiles() {
	var e Payment
	_ = e.Id
	_ = e.BatchId
	_ = e.Amount
	_ = e.Fee
}
`
	if err := os.WriteFile(filepath.Join(dir, "check_test.go"), []byte(checkCode), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}

	cmd = exec.Command("go", "build", "./...")
	cmd.Dir = dir
	out, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go build failed: %v\n%s", err, out)
	}
}
