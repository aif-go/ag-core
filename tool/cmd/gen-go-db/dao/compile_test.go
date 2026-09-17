package dao

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/model"
)

// compileGreenCases 编译矩阵关键样例（预期全绿）：有 SelfQueries（含分页）+ 无 SelfQueries（§5.2）。
var compileGreenCases = []struct {
	name string
	yaml string
}{
	{"tm_teacher", "../repository/yaml/tm_teacher.yaml"},
	{"tbl_3ds_request", "../repository/yaml/tbl_3ds_request.yaml"},
	{"tm_no", "../repository/yaml/tm_no.yaml"},
}

// compileRedCases 已知编译失败用例，登记 CHG-07（预期红，冻结项）：
// generateWithMethods 未处理 in @XxxSlice 切片参数，生成物 With 方法赋值类型不匹配。
var compileRedCases = []struct {
	name string
	yaml string
}{
	{"TM_MEDIA_ACT", "../TM_MEDIA_ACT.yaml"},
}

// findAgCoreRoot 自测试目录向上查找 ag-core 根目录（以 go.work 为标志）。
func findAgCoreRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatal("未找到 ag-core 根目录（go.work）")
	return ""
}

// generateCompileSet 将一组样例 × dbType 三态生成到临时模块的 repository/ 目录树。
func generateCompileSet(t *testing.T, genRoot string, cases []struct {
	name string
	yaml string
}) {
	t.Helper()
	for _, dc := range dbTypeCases {
		for _, cc := range cases {
			caseName := strings.ToLower(cc.name) + "_" + dc.kind
			repoDir := filepath.Join(genRoot, caseName, "repository")
			tableDatas, err := YAMLParser(cc.yaml)
			if err != nil {
				t.Fatalf("解析 %s 失败: %v", cc.yaml, err)
			}
			if len(tableDatas) == 0 {
				t.Fatalf("%s 未解析出表数据", cc.yaml)
			}
			for _, tableData := range tableDatas {
				modelDir := filepath.Join(repoDir, "model")
				daoDir := filepath.Join(repoDir, "dao")
				for _, dir := range []string{modelDir, daoDir} {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
				}
				modelPath := filepath.Join(modelDir, strings.ToLower(tableData.TableName)+"_model.go")
				if err := model.GenerateModel(tableData, modelPath); err != nil {
					t.Fatalf("生成 %s 模型失败: %v", tableData.TableName, err)
				}
				if err := GenerateDAO(tableData, daoDir, "compilegen/"+caseName, dc.dbType); err != nil {
					t.Fatalf("生成 %s DAO 失败: %v", tableData.TableName, err)
				}
			}
		}
	}
}

// setupCompileModule 写临时模块 go.mod 与临时 go.work，返回 go build 命令。
func setupCompileModule(t *testing.T, genRoot, goWorkPath, agcoreRoot string) *exec.Cmd {
	t.Helper()
	goMod := "module compilegen\n\ngo 1.24.8\n\nrequire (\n\tgithub.com/shopspring/decimal v1.4.0\n\tgorm.io/gorm v1.31.1\n\tgopkg.in/yaml.v2 v2.4.0\n)\n"
	if err := os.WriteFile(filepath.Join(genRoot, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}
	useDirs := []string{
		filepath.ToSlash(genRoot),
		filepath.ToSlash(agcoreRoot),
		filepath.ToSlash(filepath.Join(agcoreRoot, "contribute", "agdb")),
	}
	goWork := "go 1.24.8\n\nuse (\n\t" + strings.Join(useDirs, "\n\t") + "\n)\n"
	if err := os.WriteFile(goWorkPath, []byte(goWork), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "build", "compilegen/...")
	cmd.Dir = genRoot
	cmd.Env = append(os.Environ(), "GOWORK="+goWorkPath)
	return cmd
}

// TestGeneratedCodeCompiles L2 编译等价：预期绿组生成物可编译；预期红组（CHG-07）必须保持编译失败。
func TestGeneratedCodeCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("short 模式跳过编译矩阵")
	}
	agcoreRoot := findAgCoreRoot(t)
	tmp := t.TempDir()

	greenRoot := filepath.Join(tmp, "green")
	generateCompileSet(t, greenRoot, compileGreenCases)
	greenCmd := setupCompileModule(t, greenRoot, filepath.Join(tmp, "green.go.work"), agcoreRoot)
	if out, err := greenCmd.CombinedOutput(); err != nil {
		t.Fatalf("生成物编译失败（预期全绿组）: %v\n%s", err, string(out))
	}

	// CHG-04 断言：新旧方法名并存（Deprecated 委托 + 编译通过）
	daoFile := filepath.Join(greenRoot, "tm_teacher_nodbtype", "repository", "dao", "tm_teacher_dao.go")
	codeBytes, readErr := os.ReadFile(daoFile)
	if readErr != nil {
		t.Fatal(readErr)
	}
	for _, want := range []string{
		"UpdateByPrimaryKeyIgnoreZeroValCols",
		"UpdateByPrimaryKeyIngoreZeroValCols", // 旧名委托保留
		"Deprecated:",
	} {
		if !strings.Contains(string(codeBytes), want) {
			t.Errorf("生成物缺少期望标识 %q（CHG-04 双名并存）", want)
		}
	}

	redRoot := filepath.Join(tmp, "red")
	generateCompileSet(t, redRoot, compileRedCases)
	redCmd := setupCompileModule(t, redRoot, filepath.Join(tmp, "red.go.work"), agcoreRoot)
	_, err := redCmd.CombinedOutput()
	if err == nil {
		t.Fatal("预期红组（CHG-07 冻结项）编译通过：缺陷可能已修复，请评审后解冻并更新本用例")
	}
}
