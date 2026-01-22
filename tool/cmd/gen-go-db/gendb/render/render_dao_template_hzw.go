package render

import (
	"ag-core/tool/cmd/gen-go-db/gendb/render/templates"
	"bytes"
	"fmt"
	"go/format"

	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// RenderDaoTemplate
// 根据dao的html模板渲染出一个新的Dao.go文件
func RenderDaoTemplate_hzw(targetPath string, tableData *TableData) error {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	tmpl, err := template.New("dao.tmpl").Funcs(funcMap).Parse(templates.DaoTemplate)
	if err != nil {
		return err
	}

	fileName := tableData.ObjectName + "Dao.go"

	var buf bytes.Buffer
	// 渲染模板并写入文件
	// err = tmpl.Execute(file, tableData)
	err = tmpl.Execute(&buf, tableData)
	if err != nil {
		return fmt.Errorf("渲染模板 %s 失败: %w", fileName, err)
	}

	content := buf.String()
	formated, err := format.Source([]byte(content))
	if err != nil {
		return fmt.Errorf("格式化代码失败: %w", err)
	}

	fn := filepath.Join(targetPath, fileName)
	err = os.MkdirAll(filepath.Dir(fn), 0o755)
	if err == nil {
		err = os.WriteFile(fn, formated, 0o644)
	}
	if err != nil {
		return fmt.Errorf("写入文件 %s 失败: %w", fileName, err)
	}

	log.Println(fileName, " generated successfully")
	return nil
}
