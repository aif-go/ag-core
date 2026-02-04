package model

import (
	"ag-core/tool/cmd/new-gen-db/table"
	"fmt"
	"html/template"
	"os"
	"strings"
)

// GenerateModel 生成模型文件
func GenerateModel(data *table.TableData, outputPath string) error {
	// 创建模板并添加自定义函数
	tmpl := template.New("model").Funcs(template.FuncMap{
		"toLower":     strings.ToLower,
		"toCamelCase": toCamelCase,
		"inArray":     inArray,
	})

	// 解析模板
	tmpl, err := tmpl.Parse(ModelTemplate)
	if err != nil {
		return fmt.Errorf("解析模板失败: %v", err)
	}

	// 渲染模板到内存
	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return fmt.Errorf("渲染模板失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(outputPath, []byte(buf.String()), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// inArray 辅助函数，检查字段是否在查询字段中
func inArray(query table.QueryData, fieldName string) bool {
	for _, field := range query.Fields {
		if strings.EqualFold(field, fieldName) {
			return true
		}
	}
	return false
}
