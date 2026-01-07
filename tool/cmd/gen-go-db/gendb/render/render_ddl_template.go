package render

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func RenerDDLTemplate(config *AGInfraStructrueConfig,tableData *TableData) error {

	// 定义自定义函数
	// funcMap := template.FuncMap{
	// 	"unescaped": func(s string) template.HTML {
	// 		return template.HTML(s)
	// 	},
	// }

	// TODO 公共操作,线程安全问题
	// if _, err := os.Stat(targetPath); os.IsNotExist(err) {
	// 	error := os.MkdirAll(targetPath, 0755)
	// 	if error != nil {
	// 		log.Fatal("创建目录entity失败", error)
	// 	}
	// }

	templateName:=""
	switch  config.DbType {
	case MY_SQL:
		templateName="ddl_table_mysql.tmpl"
	case DB2:
		templateName="ddl_table_db2.tmpl"
	default:
		return fmt.Errorf("不支持的db类型,当前仅支持db2和mysql")
	}
	// 加载模板文件
	tmpl, err := GetTemplate(templateName, nil)
	if err != nil {
		return err
	}

	fileName := tableData.ObjectName + ".sql"

	// 创建输出文件
	file, err := os.Create(filepath.Join(config.SqlPath, fileName))
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
	}

	defer file.Close()

	// 渲染模板并写入文件
	err = tmpl.Execute(file, tableData)
	if err != nil {
		return fmt.Errorf("渲染模板 %s 失败: %w", file.Name(), err)
	}
	log.Println("Go file ", fileName, " generated successfully")
	return nil
}