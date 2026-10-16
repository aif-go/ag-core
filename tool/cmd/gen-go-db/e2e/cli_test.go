package e2e

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// gendbBin TestMain 编译好的 gendb 二进制路径，全部用例共享。
var gendbBin string

// TestMain 整包编译一次 gendb 二进制，避免每用例重复编译。
func TestMain(m *testing.M) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	moduleRoot := ""
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
			if b, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil && strings.HasPrefix(string(b), "module github.com/aif-go/ag-core/tool/cmd/gen-go-db") {
				moduleRoot = dir
				break
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if moduleRoot == "" {
		fmt.Fprintln(os.Stderr, "未找到 gen-go-db 模块根目录")
		os.Exit(1)
	}
	bin := filepath.Join(os.TempDir(), "gendb-e2e.exe")
	cmd := exec.Command("go", "build", "-o", bin, ".")
	cmd.Dir = moduleRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "编译 gendb 失败: %v\n%s", err, string(out))
		os.Exit(1)
	}
	gendbBin = bin
	os.Exit(m.Run())
}

// buildGendbBinary 返回 TestMain 编译好的二进制路径与模块根目录。
func buildGendbBinary(t *testing.T) (string, string) {
	t.Helper()
	if gendbBin == "" {
		t.Fatal("TestMain 未完成二进制编译")
	}
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "main.go")); err == nil {
			if b, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil && strings.HasPrefix(string(b), "module github.com/aif-go/ag-core/tool/cmd/gen-go-db") {
				return gendbBin, dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("未找到 gen-go-db 模块根目录")
	return "", ""
}

// runGendb 执行子命令，返回退出码与合并输出。
func runGendb(t *testing.T, bin string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.CombinedOutput()
	if err == nil {
		return 0, string(out)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), string(out)
	}
	t.Fatalf("执行 gendb 失败: %v\n%s", err, string(out))
	return -1, string(out)
}

// runGendbStdout 执行子命令，仅取 stdout（--json 契约验证用）；stderr 被忽略。
func runGendbStdout(t *testing.T, bin string, args ...string) (int, string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	out, err := cmd.Output()
	if err == nil {
		return 0, string(out)
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		return exitErr.ExitCode(), string(out)
	}
	t.Fatalf("执行 gendb 失败: %v\n%s", err, string(out))
	return -1, string(out)
}

// TestE2E_Db_MissingArgs db 子命令必填校验语义不变（L4）。
func TestE2E_Db_MissingArgs(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	code, out := runGendb(t, bin, "db")
	if code != 1 {
		t.Fatalf("缺参应退出码 1，实际 %d，输出:\n%s", code, out)
	}
	if !strings.Contains(out, "错误：输入参数不能为空") {
		t.Errorf("缺参错误语义变化，输出:\n%s", out)
	}
}

// TestE2E_Yaml_MissingArgs yaml 子命令必填校验语义不变（L4）。
func TestE2E_Yaml_MissingArgs(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	code, out := runGendb(t, bin, "yaml")
	if code != 1 {
		t.Fatalf("缺参应退出码 1，实际 %d，输出:\n%s", code, out)
	}
	if !strings.Contains(out, "错误：输入参数不能为空") {
		t.Errorf("缺参错误语义变化，输出:\n%s", out)
	}
}

// TestE2E_Sheet_MissingArgs sheet 子命令必填校验语义不变（L4）。
func TestE2E_Sheet_MissingArgs(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	code, out := runGendb(t, bin, "sheet")
	if code != 1 {
		t.Fatalf("缺参应退出码 1，实际 %d，输出:\n%s", code, out)
	}
	if !strings.Contains(out, "错误：输入参数不能为空") {
		t.Errorf("缺参错误语义变化，输出:\n%s", out)
	}
}

// TestE2E_Yaml_TestMode yaml 子命令测试模式：输出目录拼接 repository/yaml（L4）。
func TestE2E_Yaml_TestMode(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	outDir := t.TempDir()
	code, out := runGendb(t, bin, "yaml", "-t", "-o", outDir)
	if code != 0 {
		t.Fatalf("yaml -t 应成功，退出码 %d，输出:\n%s", code, out)
	}
	yamlDir := filepath.Join(outDir, "repository", "yaml")
	entries, err := os.ReadDir(yamlDir)
	if err != nil {
		t.Fatalf("输出目录结构变化，缺少 %s: %v", yamlDir, err)
	}
	if len(entries) == 0 {
		t.Fatal("yaml -t 未生成任何 YAML 文件")
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".yaml") {
			t.Errorf("repository/yaml 下出现非 YAML 文件: %s", e.Name())
		}
	}
}

// TestE2E_Db_TmNo db 子命令：无 SelfQueries 表的文件清单与目录结构（L4）。
func TestE2E_Db_TmNo(t *testing.T) {
	bin, moduleRoot := buildGendbBinary(t)
	outDir := t.TempDir()
	code, out := runGendb(t, bin, "db", "-i", filepath.Join(moduleRoot, "repository", "yaml", "tm_no.yaml"), "-o", outDir, "-m", "example.com/demo")
	if code != 0 {
		t.Fatalf("db 应成功，退出码 %d，输出:\n%s", code, out)
	}
	for _, want := range []string{
		filepath.Join("repository", "dao", "tm_no_dao.go"),
		filepath.Join("repository", "dao", "tm_no_constant.go"),
		filepath.Join("repository", "model", "tm_no_model.go"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, want)); err != nil {
			t.Errorf("缺少生成文件 %s", want)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "repository", "dao", "tm_no_namingsql.go")); !os.IsNotExist(err) {
		t.Error("无 SelfQueries 表不应生成 namingsql 文件")
	}
}

// TestE2E_Db_TmTeacherMysql db 子命令：-d mysql 文件清单（L4，含 CHG-05 相关语义：单库只生成对应 dbtype 文件）。
func TestE2E_Db_TmTeacherMysql(t *testing.T) {
	bin, moduleRoot := buildGendbBinary(t)
	outDir := t.TempDir()
	code, out := runGendb(t, bin, "db", "-i", filepath.Join(moduleRoot, "repository", "yaml", "tm_teacher.yaml"), "-o", outDir, "-m", "example.com/demo", "-d", "mysql")
	if code != 0 {
		t.Fatalf("db 应成功，退出码 %d，输出:\n%s", code, out)
	}
	for _, want := range []string{
		filepath.Join("repository", "dao", "tm_teacher_dao.go"),
		filepath.Join("repository", "dao", "tm_teacher_constant.go"),
		filepath.Join("repository", "dao", "tm_teacher_namingsql.go"),
		filepath.Join("repository", "dao", "mysql_tm_teacher_namingsql.go"),
		filepath.Join("repository", "model", "tm_teacher_model.go"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, want)); err != nil {
			t.Errorf("缺少生成文件 %s", want)
		}
	}
	if _, err := os.Stat(filepath.Join(outDir, "repository", "dao", "db2_tm_teacher_namingsql.go")); !os.IsNotExist(err) {
		t.Error("-d mysql 不应生成 db2 前缀文件")
	}
}

// TestE2E_Db_BadInput db 子命令错误返回语义不变（L4）。
func TestE2E_Db_BadInput(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	code, out := runGendb(t, bin, "db", "-i", filepath.Join(t.TempDir(), "not_exist.yaml"), "-o", t.TempDir())
	if code != 1 {
		t.Fatalf("非法输入应退出码 1，实际 %d，输出:\n%s", code, out)
	}
	if !strings.Contains(out, "生成DAO文件失败") {
		t.Errorf("错误返回语义变化，输出:\n%s", out)
	}
}

// TestE2E_Db_InvalidDbType CHG-05：非法 -d 值生成前报错，不再产出编译不过的产物。
func TestE2E_Db_InvalidDbType(t *testing.T) {
	bin, moduleRoot := buildGendbBinary(t)
	outDir := t.TempDir()
	code, out := runGendb(t, bin, "db", "-i", filepath.Join(moduleRoot, "repository", "yaml", "tm_no.yaml"), "-o", outDir, "-m", "example.com/demo", "-d", "oracle")
	if code != 1 {
		t.Fatalf("非法 dbType 应退出码 1，实际 %d，输出:\n%s", code, out)
	}
	if !strings.Contains(out, "非法的数据库类型") {
		t.Errorf("应输出 dbType 校验错误，输出:\n%s", out)
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("校验失败后不应产生任何输出文件，实际: %v", entries)
	}
}

// buildSheetInput 构造 sheet 子命令输入 xlsx。
func buildSheetInput(t *testing.T, path string) {
	t.Helper()
	f := excelize.NewFile()
	sheet := "demo_table"
	f.NewSheet(sheet)
	f.SetCellValue(sheet, "A1", "id")
	f.SetCellValue(sheet, "A2", "1")
	f.SetCellValue(sheet, "A3", "自定义脚本名字")
	f.SetCellValue(sheet, "A4", "2")
	f.DeleteSheet("Sheet1")
	if err := f.SaveAs(path); err != nil {
		t.Fatal(err)
	}
}

// TestE2E_Sheet_Split sheet 子命令：拆分输出文件落盘（L4）。
func TestE2E_Sheet_Split(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	inDir := t.TempDir()
	input := filepath.Join(inDir, "in.xlsx")
	buildSheetInput(t, input)
	outDir := t.TempDir()
	code, out := runGendb(t, bin, "sheet", "-i", input, "-o", outDir, "-k", "自定义脚本名字")
	if code != 0 {
		t.Fatalf("sheet 应成功，退出码 %d，输出:\n%s", code, out)
	}
	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "in_") && strings.HasSuffix(e.Name(), ".xlsx") {
			found = true
		}
	}
	if !found {
		t.Errorf("输出目录缺少 in_<日期>.xlsx，实际: %v", entries)
	}
}

// TestE2E_Db_JSONSuccess db --json 成功契约：stdout 为单一 JSON、ok=true、产物落盘、退出码 0。
func TestE2E_Db_JSONSuccess(t *testing.T) {
	bin, moduleRoot := buildGendbBinary(t)
	outDir := t.TempDir()
	code, stdout := runGendbStdout(t, bin, "db", "-i", filepath.Join(moduleRoot, "repository", "yaml", "tm_no.yaml"), "-o", outDir, "-m", "example.com/demo", "--json")
	if code != 0 {
		t.Fatalf("db --json 应成功，退出码 %d，stdout:\n%s", code, stdout)
	}
	var got struct {
		OK bool `json:"ok"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout 不是单一合法 JSON: %v\n%s", err, stdout)
	}
	if !got.OK {
		t.Errorf("ok 应为 true, stdout: %s", stdout)
	}
	for _, want := range []string{
		filepath.Join("repository", "dao", "tm_no_dao.go"),
		filepath.Join("repository", "dao", "tm_no_constant.go"),
		filepath.Join("repository", "model", "tm_no_model.go"),
	} {
		if _, err := os.Stat(filepath.Join(outDir, want)); err != nil {
			t.Errorf("缺少生成文件 %s", want)
		}
	}
}

// TestE2E_Db_JSONUnknownType db --json 失败契约：未识别类型 → 退出码 1、ok=false、
// error.phase=parse、issues 含列与原始类型。
func TestE2E_Db_JSONUnknownType(t *testing.T) {
	bin, _ := buildGendbBinary(t)
	bad := filepath.Join(t.TempDir(), "bad.yaml")
	content := `table_name: tm_bad
columns:
- name: a
  type: bytea
`
	if err := os.WriteFile(bad, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	code, stdout := runGendbStdout(t, bin, "db", "-i", bad, "-o", t.TempDir(), "-m", "example.com/demo", "--json")
	if code != 1 {
		t.Fatalf("db --json 类型非法应退出码 1，实际 %d，stdout:\n%s", code, stdout)
	}
	var got struct {
		OK    bool `json:"ok"`
		Error struct {
			Phase  string `json:"phase"`
			Table  string `json:"table"`
			Issues []struct {
				Column string `json:"column"`
				Type   string `json:"type"`
			} `json:"issues"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout 不是单一合法 JSON: %v\n%s", err, stdout)
	}
	if got.OK {
		t.Errorf("ok 应为 false, stdout: %s", stdout)
	}
	if got.Error.Phase != "parse" {
		t.Errorf("phase = %q; want parse", got.Error.Phase)
	}
	if len(got.Error.Issues) != 1 || got.Error.Issues[0].Type != "bytea" || got.Error.Issues[0].Column != "a" {
		t.Errorf("issues 不符: %+v", got.Error.Issues)
	}
}

// TestE2E_Db_DryRun db --dry-run 契约：退出码 0、输出目录零产物、摘要列出计划文件。
func TestE2E_Db_DryRun(t *testing.T) {
	bin, moduleRoot := buildGendbBinary(t)
	outDir := t.TempDir()
	code, out := runGendb(t, bin, "db", "-i", filepath.Join(moduleRoot, "repository", "yaml", "tm_no.yaml"), "-o", outDir, "-m", "example.com/demo", "--dry-run")
	if code != 0 {
		t.Fatalf("db --dry-run 应成功，退出码 %d，输出:\n%s", code, out)
	}
	if _, err := os.Stat(filepath.Join(outDir, "repository")); !os.IsNotExist(err) {
		t.Error("dry-run 不应在输出目录产生任何产物")
	}
	if !strings.Contains(out, "repository/model/tm_no_model.go") {
		t.Errorf("dry-run 应列出计划文件，输出:\n%s", out)
	}
}
