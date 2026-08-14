package dao

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// normalizeNewlines 归一化换行符，消除 Windows CRLF 与模板 LF 的差异。
func normalizeNewlines(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}

// extractFindByStruct 提取 FindByStruct 方法体（含签名与注释），用于与生成产物逐字节比对。
func extractFindByStruct(code string) string {
	start := strings.Index(code, "// FindByStruct 根据实体查询")
	if start < 0 {
		return ""
	}
	end := strings.Index(code[start:], "\n// FindByCustomerRule")
	if end < 0 {
		return code[start:]
	}
	return code[start : start+end]
}

// TestGeneratedDAOConsistency 校验 DAO 模板输出与已生成产物逐字节一致，
// 防止模板改动后产物未同步（对应 FindByStruct 索引判定简化改造的闭环验证）。
func TestGeneratedDAOConsistency(t *testing.T) {
	moduleName := "github.com/aif-go/ag-core/tool/cmd/gen-go-db"

	cases := []struct {
		yamlFile string
		daoFile  string
	}{
		{"../repository/yaml/tbl_3ds_request.yaml", "../repository/dao/tbl_3ds_request_dao.go"},
		{"../repository/yaml/tm_teacher.yaml", "../repository/dao/tm_teacher_dao.go"},
	}

	for _, c := range cases {
		t.Run(filepath.Base(c.yamlFile), func(t *testing.T) {
			tableDatas, err := YAMLParser(c.yamlFile)
			if err != nil {
				t.Fatalf("解析 %s 失败: %v", c.yamlFile, err)
			}
			if len(tableDatas) == 0 {
				t.Fatalf("%s 未解析出表数据", c.yamlFile)
			}
			tableData := tableDatas[0]
			tableData.ModuleName = moduleName

			// 仅比对 FindByStruct 方法体：该方法是本 change 的生成范围，
			// 且样例 DAO 的 FindByCustomerRule case 顺序存在既有历史差异，不适合整文件比对。
			got := extractFindByStruct(normalizeNewlines(GetDaoTemplate(tableData)))
			wantBytes, err := os.ReadFile(c.daoFile)
			if err != nil {
				t.Fatalf("读取产物 %s 失败: %v", c.daoFile, err)
			}
			want := extractFindByStruct(normalizeNewlines(string(wantBytes)))
			if got != want {
				gotLines := strings.Split(got, "\n")
				wantLines := strings.Split(want, "\n")
				diff := ""
				for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
					var g, w string
					if i < len(gotLines) {
						g = gotLines[i]
					}
					if i < len(wantLines) {
						w = wantLines[i]
					}
					if g != w {
						diff += fmt.Sprintf("第%d行\n  模板: %q\n  产物: %q\n", i+1, g, w)
						break
					}
				}
				t.Errorf("模板输出与产物不一致: %s\n首个差异:\n%s", c.daoFile, diff)
			}
		})
	}
}
