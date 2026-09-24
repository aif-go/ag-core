package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// optimisticlockYaml 乐观锁列样例 YAML（锁列改名防硬编码：非 jpa_version 命名）。
const optimisticlockYaml = `table_name: tm_lock_demo
columns:
- name: id
  type: int64
- name: version_no
  type: int64
  tag: ///@optimisticlock
- name: name
  type: string
  length: "20"
`

// TestParseYAML_OptimisticLock 乐观锁 tag 解析：IsOptimisticLock、GoType 固定、import 注入。
func TestParseYAML_OptimisticLock(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "lock.yaml")
	if err := os.WriteFile(yamlPath, []byte(optimisticlockYaml), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 意外报错: %v", err)
	}

	var lockCol *table.ColumnData
	for i := range data.Columns {
		if data.Columns[i].IsOptimisticLock {
			lockCol = &data.Columns[i]
			break
		}
	}
	if lockCol == nil {
		t.Fatal("未解析出 IsOptimisticLock 列")
	}

	// 锁列改名（version_no，非 jpa_version）同样生效——防字段名硬编码回归
	if lockCol.Name != "version_no" {
		t.Errorf("lockCol.Name = %q; want version_no", lockCol.Name)
	}
	if lockCol.GoType != "optimisticlock.Version" {
		t.Errorf("GoType = %q; want optimisticlock.Version（tag 声明的 type 被覆盖）", lockCol.GoType)
	}

	found := false
	for _, pkg := range data.ModelTemplateData.ImportPackages {
		if pkg == "gorm.io/plugin/optimisticlock" {
			found = true
		}
	}
	if !found {
		t.Errorf("未注入 gorm.io/plugin/optimisticlock import, got: %v", data.ModelTemplateData.ImportPackages)
	}

	// 普通列不受影响
	for _, col := range data.Columns {
		if col.Name == "id" && (col.IsOptimisticLock || col.GoType != "int64") {
			t.Errorf("普通列被误标: id GoType=%q IsOptimisticLock=%v", col.GoType, col.IsOptimisticLock)
		}
	}
}

// TestParseYAML_OptimisticLock_JavaVersionCoexist ///@optimisticlock 与其他 tag 并存（; 分隔）
func TestParseYAML_OptimisticLock_JavaVersionCoexist(t *testing.T) {
	yamlContent := `table_name: tm_lock_coexist
columns:
- name: id
  type: int64
- name: jpa_version
  type: int64
  tag: ///@javaVersion;///@optimisticlock
`
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "lock_coexist.yaml")
	if err := os.WriteFile(yamlPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}

	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 意外报错: %v", err)
	}
	col := data.Columns[1]

	if !col.IsOptimisticLock || !col.IsJavaVersion {
		t.Fatalf("双 tag 应同时生效: IsOptimisticLock=%v IsJavaVersion=%v", col.IsOptimisticLock, col.IsJavaVersion)
	}
	if col.GoType != "optimisticlock.Version" {
		t.Errorf("GoType = %q; want optimisticlock.Version（乐观锁声明优先）", col.GoType)
	}

	imported := false
	for _, pkg := range data.ModelTemplateData.ImportPackages {
		if pkg == "gorm.io/plugin/optimisticlock" {
			imported = true
		}
	}
	if !imported {
		t.Error("双 tag 场景未注入 optimisticlock import")
	}
}

// TestGetZeroCheck_OptimisticVersion 乐观锁列零值判断以 Valid 为准（值 0 属合法历史版本）。
func TestGetZeroCheck_OptimisticVersion(t *testing.T) {
	col := table.ColumnData{Name: "version_no", JsonTag: "VersionNo", GoType: "optimisticlock.Version"}
	check := getZeroCheck("tmLockDemo", col)
	if check != "!tmLockDemo.VersionNo.Valid" {
		t.Errorf("getZeroCheck = %q; want !tmLockDemo.VersionNo.Valid", check)
	}
}

// TestIsSpecialColumn_OptimisticLock 锁列归入特殊列（剔除名单保护）。
func TestIsSpecialColumn_OptimisticLock(t *testing.T) {
	col := table.ColumnData{IsOptimisticLock: true}
	if !isSpecialColumn(col) {
		t.Error("乐观锁列应为特殊列（exclude 出零值剔除名单）")
	}
}

// TestFieldChecksCode_OptimisticLock 渲染 ListZeroValueCols：锁列走 Valid 判断且归为特殊列（渲染级断言）。
func TestFieldChecksCode_OptimisticLock(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "lock.yaml")
	if err := os.WriteFile(yamlPath, []byte(optimisticlockYaml), 0644); err != nil {
		t.Fatalf("写入YAML文件失败: %v", err)
	}
	data, err := ParseYAML(yamlPath, "test-module")
	if err != nil {
		t.Fatalf("ParseYAML 意外报错: %v", err)
	}

	code := GetModelTemplate(data)

	if !strings.Contains(code, "!tmLockDemo.VersionNo.Valid") {
		t.Errorf("FieldChecks 渲染未使用 Valid 判断:\n%s", code)
	}
	if !strings.Contains(code, "用于乐观锁") {
		t.Errorf("锁列应归类特殊列（用于乐观锁）:\n%s", code)
	}
}
