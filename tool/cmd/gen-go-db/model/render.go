package model

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
	"go/format"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

//go:embed templates/*.tmpl
var modelTemplates embed.FS

// renderModelTemplate 渲染指定模板（阶段 3 不做 gofmt，见方案 §7.4）。
func renderModelTemplate(name string, data any) (string, error) {
	tmpl, err := template.ParseFS(modelTemplates, "templates/*.tmpl")
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, name, data); err != nil {
		return "", err
	}
	// 阶段 4（CHG-01）：生成物统一 gofmt；格式化失败（如 CHG-07 缺陷产物）回退原文
	if formatted, ferr := format.Source(buf.Bytes()); ferr == nil {
		return string(formatted), nil
	}
	return buf.String(), nil
}

// GetModelTemplate 获取Model模板代码
func GetModelTemplate(tableData *table.TableData) string {
	data := BuildModelTemplateData(tableData)
	code, err := renderModelTemplate("model.go.tmpl", data)
	if err != nil {
		panic(fmt.Errorf("渲染 model 模板失败: %w", err))
	}
	return code
}
