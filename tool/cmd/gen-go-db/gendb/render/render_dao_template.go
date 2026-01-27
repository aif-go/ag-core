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
func RenderDaoTemplate(targetPath string, tableData *TableData) error {
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

// func RenderDaoTemplate(targetPath string, tableData *TableData) error {
// 	funcMap := template.FuncMap{
// 		"ToLower": strings.ToLower,
// 	}
// 	// 加载模板文件
// 	tmpl, err := GetTemplate("dao.tmpl", funcMap)

// 	if err != nil {
// 		return err
// 	}

// 	fileName := tableData.ObjectName + "Dao.go"
// 	// 创建输出文件
// 	file, err := os.Create(filepath.Join(targetPath, fileName))

// 	if err != nil {
// 		return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
// 	}
// 	defer file.Close()
// 	// 渲染模板并写入文件
// 	err = tmpl.Execute(file, tableData)
// 	if err != nil {
// 		return fmt.Errorf("渲染模板 %s 失败: %w", fileName, err)
// 	}
// 	log.Println(fileName, " generated successfully")
// 	return nil
// }

// RenderNamingSqlConstant 生成namingsql的自定义sql的常量
func RenderNamingSqlConstant(config *AGInfraStructrueConfig, tableData *TableData) error {
	// 定义自定义函数,用来解决引号被转成ascii码的问题
	funcMap := template.FuncMap{
		"unescaped": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	// 加载模板文件
	tmpl, err := GetTemplate("naming_sql.tmpl", funcMap)

	if err != nil {
		return err
	}

	// 保证即使切换db的时候,只是新增内容，但是对应的数据是不会变的
	targetPath := config.DaoPath
	for _, support := range config.SupportDB {
		
		dbType := strings.ToUpper(support)
		fileName := dbType + "_" + tableData.ObjectName + "_NamingSql.go"
		// 为nil的时候如何处理，此时全部生成
		if  config.DbType == "" {
			_, err := os.Stat(targetPath + fileName)
			// 说明文件存在
			if err == nil {
				continue
			}

			// 如果当前生成的db类型和支持列表中的不一致，按照既定逻辑处理
			// 1. 对应的db文件不存在，则生成新的文件按照配置命名sql，但是sql内容为空，存在则跳过处理
			if os.IsExist(err) {
				continue
			}			
		}else if tableData.DbType != dbType {
			continue
		}
		nullData := &TableData{
				DbType:           dbType,
				ObjectName:       tableData.ObjectName,
				RNamingSqlList:   []*NamingSqlTemplate{},
				CUDNamingSqlList: []*NamingSqlTemplate{},
				PackageName:      tableData.PackageName,
			}

			for _, namingsql := range tableData.CUDNamingSqlList {

				sql := namingsql.NamingSql
				if value, ok := tableData.NamingSqlMap[dbType+"@"+namingsql.MethodName]; ok {
					sql = value.NamingSql
				}
				nullData.CUDNamingSqlList = append(nullData.CUDNamingSqlList,
					&NamingSqlTemplate{NamingSql: sql, MethodName: namingsql.MethodName, PageCountSql: namingsql.PageCountSql})
			}

			for _, namingsql := range tableData.RNamingSqlList {
				sql := namingsql.NamingSql
				if value, ok := tableData.NamingSqlMap[dbType+"@"+namingsql.MethodName]; ok {
					sql = value.NamingSql
				}
				nullData.RNamingSqlList = append(nullData.RNamingSqlList,
					&NamingSqlTemplate{NamingSql: sql, MethodName: namingsql.MethodName, PageCountSql: namingsql.PageCountSql})
			}
		tableData = nullData
		// 创建输出文件
		file, err := os.Create(filepath.Join(targetPath, fileName))
		if err != nil {
			return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
		}
		defer file.Close()
		// 渲染模板并写入文件
		err = tmpl.Execute(file, tableData)
		if err != nil {
			return fmt.Errorf("渲染模板 %s 失败: %w", fileName, err)
		}
		log.Println(fileName, " generated successfully")
	}
	return nil
}

type TableNamingInitMethodData struct {
	ObjectName  string
	DbTypeSlice []string
}

// RenderNamingSqlConstant 生成namingsql的自定义sql的常量
func RenderTableNamingSql(config *AGInfraStructrueConfig, tableData *TableData) error {
	// 加载模板文件
	tmpl, err := GetTemplate("table_naming_sql.tmpl", nil)

	if err != nil {
		return err
	}

	// 保证即使切换db的时候,只是新增内容，但是对应的数据是不会变的
	targetPath := config.DaoPath
	fileName := tableData.ObjectName + "_NamingSql.go"
	// 创建输出文件
	file, err := os.Create(filepath.Join(targetPath, fileName))
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
	}
	defer file.Close()

	renderData := []string{}
	if config.DbType == "" {
	for _, dbType := range config.SupportDB {
		renderData = append(renderData, strings.ToUpper(dbType))
	}
	} else {	
		renderData = append(renderData, strings.ToUpper(config.DbType))
	}
	data := &TableNamingInitMethodData{
		ObjectName:  tableData.ObjectName,
		DbTypeSlice: renderData,
	}
	// 渲染模板并写入文件
	err = tmpl.Execute(file, data)
	if err != nil {
		return fmt.Errorf("渲染模板 %s 失败: %w", fileName, err)
	}
	log.Println(fileName, " generated successfully")
	return nil
}
