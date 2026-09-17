package dao

import (
	"bytes"
	"embed"
	"fmt"
	"text/template"
	"go/format"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

//go:embed templates/*.tmpl
var daoTemplates embed.FS

// renderDaoTemplate 渲染指定模板（阶段 3 不做 gofmt，见方案 §7.4）。
func renderDaoTemplate(name string, data any) (string, error) {
	tmpl, err := template.ParseFS(daoTemplates, "templates/*.tmpl")
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

// GetDaoTemplate 获取DAO模板代码
func GetDaoTemplate(tableData *table.TableData) string {
	data := &DaoTemplateData{TableData: tableData}
	code, err := renderDaoTemplate("dao.go.tmpl", data)
	if err != nil {
		panic(fmt.Errorf("渲染 dao 模板失败: %w", err))
	}
	return code
}

// GetConstantTemplate 获取常量模板代码
func GetConstantTemplate(tableData *table.TableData) string {
	data := BuildConstantData(tableData)
	code, err := renderDaoTemplate("constant.go.tmpl", data)
	if err != nil {
		panic(fmt.Errorf("渲染 constant 模板失败: %w", err))
	}
	return code
}

// GetNamingSqlTemplate 获取命名SQL模板代码
func GetNamingSqlTemplate(tableData *table.TableData, dbType string) string {
	data := BuildNamingSqlData(tableData, dbType)
	code, err := renderDaoTemplate("namingsql.go.tmpl", data)
	if err != nil {
		panic(fmt.Errorf("渲染 namingsql 模板失败: %w", err))
	}
	return code
}

// GetDBTypeNamingSqlTemplate 获取数据库类型命名SQL模板代码
func GetDBTypeNamingSqlTemplate(tableData *table.TableData, dbType string) (string, error) {
	data, err := BuildDBTypeNamingSqlData(tableData, dbType)
	if err != nil {
		return "", err
	}
	return renderDaoTemplate("dbtype_namingsql.go.tmpl", data)
}
