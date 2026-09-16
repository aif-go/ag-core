package dao

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/model"
)

// updateGolden 由 -update 标志控制，重写 golden 基线。
var updateGolden = flag.Bool("update", false, "重写 golden 基线")

// goldenModuleName 生成产物使用的模块名，与 repository/ 既有产物一致。
const goldenModuleName = "github.com/aif-go/ag-core/tool/cmd/gen-go-db"

// goldenCases 覆盖矩阵样例：7 个 YAML（§4.2）。
var goldenCases = []struct {
	name string
	yaml string
}{
	{"tm_teacher", "../repository/yaml/tm_teacher.yaml"},
	{"tm_student", "../repository/yaml/tm_student.yaml"},
	{"tm_no", "../repository/yaml/tm_no.yaml"},
	{"tm_no_index", "../repository/yaml/tm_no_index.yaml"},
	{"tm_no_primary", "../repository/yaml/tm_no_primary.yaml"},
	{"tbl_3ds_request", "../repository/yaml/tbl_3ds_request.yaml"},
	{"TM_MEDIA_ACT", "../TM_MEDIA_ACT.yaml"},
}

// dbTypeCases dbType 三态：不指定 / mysql / db2（§4.2）。
var dbTypeCases = []struct {
	kind   string
	dbType string
}{
	{"nodbtype", ""},
	{"mysql", "mysql"},
	{"db2", "db2"},
}

// normalizeNewlines 归一化换行符，消除 Windows CRLF 与模板 LF 的差异。
func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// generateGoldenCase 按 §5.1 管线生成产物：YAMLParser → model.GenerateModel → dao.GenerateDAO。
func generateGoldenCase(t *testing.T, outDir, yamlPath, dbType string) {
	t.Helper()
	tableDatas, err := YAMLParser(yamlPath)
	if err != nil {
		t.Fatalf("解析 %s 失败: %v", yamlPath, err)
	}
	if len(tableDatas) == 0 {
		t.Fatalf("%s 未解析出表数据", yamlPath)
	}
	modelDir := filepath.Join(outDir, "model")
	daoDir := filepath.Join(outDir, "dao")
	for _, dir := range []string{modelDir, daoDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, tableData := range tableDatas {
		modelPath := filepath.Join(modelDir, strings.ToLower(tableData.TableName)+"_model.go")
		if err := model.GenerateModel(tableData, modelPath); err != nil {
			t.Fatalf("生成 %s 模型失败: %v", tableData.TableName, err)
		}
		if err := GenerateDAO(tableData, daoDir, goldenModuleName, dbType); err != nil {
			t.Fatalf("生成 %s DAO 失败: %v", tableData.TableName, err)
		}
	}
}

// collectGoldenFiles 收集目录下 model/ 与 dao/ 全部文件的归一化内容，键为相对路径。
func collectGoldenFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		if !strings.HasPrefix(key, "model/") && !strings.HasPrefix(key, "dao/") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[key] = normalizeNewlines(string(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// goldenFirstDiff 返回两个文本的首个差异行描述。
func goldenFirstDiff(a, b string) string {
	aLines := strings.Split(a, "\n")
	bLines := strings.Split(b, "\n")
	for i := 0; i < len(aLines) || i < len(bLines); i++ {
		var la, lb string
		if i < len(aLines) {
			la = aLines[i]
		}
		if i < len(bLines) {
			lb = bLines[i]
		}
		if la != lb {
			return fmt.Sprintf("第%d行\n  实际: %q\n  基线: %q", i+1, la, lb)
		}
	}
	return ""
}

// TestGoldenGeneration L1 字节等价门禁：生成产物与 golden 基线逐字节比对（换行符归一化后）。
func TestGoldenGeneration(t *testing.T) {
	root := goldenRoot()
	for _, gc := range goldenCases {
		for _, dc := range dbTypeCases {
			t.Run(gc.name+"/"+dc.kind, func(t *testing.T) {
				goldenDir := filepath.Join(root, gc.name, dc.kind)
				outDir := t.TempDir()
				generateGoldenCase(t, outDir, gc.yaml, dc.dbType)
				got := collectGoldenFiles(t, outDir)
				if len(got) == 0 {
					t.Fatal("无生成文件")
				}
				if *updateGolden {
					if err := os.RemoveAll(goldenDir); err != nil {
						t.Fatal(err)
					}
					for rel, content := range got {
						dst := filepath.Join(goldenDir, filepath.FromSlash(rel))
						if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(dst, []byte(content), 0644); err != nil {
							t.Fatal(err)
						}
					}
					return
				}
				if _, err := os.Stat(goldenDir); os.IsNotExist(err) {
					t.Fatalf("golden 基线缺失: %s（先运行 go test ./dao -run TestGoldenGeneration -update 固化基线）", goldenDir)
				}
				want := collectGoldenFiles(t, goldenDir)
				for rel, wantContent := range want {
					gotContent, ok := got[rel]
					if !ok {
						t.Errorf("缺少生成文件: %s", rel)
						continue
					}
					if gotContent != wantContent {
						t.Errorf("产物与基线不一致: %s\n首个差异:\n%s", rel, goldenFirstDiff(gotContent, wantContent))
					}
				}
				for rel := range got {
					if _, ok := want[rel]; !ok {
						t.Errorf("多出生成文件: %s（基线需更新或产物漂移）", rel)
					}
				}
			})
		}
	}
}

// goldenRoot golden 基线根目录。
func goldenRoot() string {
	return filepath.Join("testdata", "golden")
}
