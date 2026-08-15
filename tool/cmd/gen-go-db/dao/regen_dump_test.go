package dao

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/utils"
)

// TestRegenDump 一次性辅助：把模板全量输出写到临时目录，供外部 diff 核对产物状态。
func TestRegenDump(t *testing.T) {
	moduleName := "github.com/aif-go/ag-core/tool/cmd/gen-go-db"
	_ = utils.ContainsIgnoreCase
	for _, yamlFile := range []string{"../repository/yaml/tm_teacher.yaml", "../repository/yaml/tbl_3ds_request.yaml"} {
		tables, err := YAMLParser(yamlFile)
		if err != nil {
			t.Fatal(err)
		}
		tables[0].ModuleName = moduleName
		code := normalizeNewlines(GetDaoTemplate(tables[0]))
		name := "regen_" + filepath.Base(yamlFile) + ".go"
		if err := os.WriteFile(filepath.Join(os.TempDir(), name), []byte(code), 0644); err != nil {
			t.Fatal(err)
		}
		t.Log("written:", filepath.Join(os.TempDir(), name))
	}
}
