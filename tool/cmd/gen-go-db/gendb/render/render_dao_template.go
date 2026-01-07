package render

import (
	"fmt"
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
	// 加载模板文件
	tmpl, err := GetTemplate("dao.tmpl", funcMap)

	if err != nil {
		return err
	}

	fileName:=tableData.ObjectName+"Dao.go"
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
	return nil
}

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
	for _,support:= range config.SupportDB {
		dbType:=strings.ToUpper(support)
		fileName:=dbType+"_"+tableData.ObjectName+"_NamingSql.go"
		if dbType != tableData.DbType{
			_,err:=os.Stat(targetPath+fileName)
			// 说明文件存在
			if err == nil{
				continue
			}

			// 如果当前生成的db类型和支持列表中的不一致，按照既定逻辑处理
			// 1. 对应的db文件不存在，则生成新的文件按照配置命名sql，但是sql内容为空，存在则跳过处理
			if os.IsExist(err){
				continue
			}

			nullData := &TableData{
				DbType: dbType,
				ObjectName:  tableData.ObjectName,
				RNamingSqlList: []*NamingSqlTemplate{},
				CUDNamingSqlList: []*NamingSqlTemplate{},
				PackageName:  tableData.PackageName,
			}
		

			for _,namingsql:=range tableData.CUDNamingSqlList {
			
				sql:=namingsql.NamingSql
				if value,ok:=tableData.NamingSqlMap[dbType+"@"+namingsql.MethodName]; ok{
					sql = value.NamingSql
				}
				nullData.CUDNamingSqlList = append(nullData.CUDNamingSqlList, 
					&NamingSqlTemplate{NamingSql: sql, MethodName: namingsql.MethodName})
			}

			for _,namingsql:=range tableData.RNamingSqlList {
				sql:=namingsql.NamingSql
				if value,ok:=tableData.NamingSqlMap[dbType+"@"+namingsql.MethodName]; ok{
					sql = value.NamingSql
				}
				nullData.RNamingSqlList = append(nullData.RNamingSqlList, 
					&NamingSqlTemplate{NamingSql:sql, MethodName: namingsql.MethodName})
			}
			tableData = nullData
		}

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


type TableNamingInitMethodData struct{

	ObjectName string
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
	fileName:=tableData.ObjectName+"_NamingSql.go"
	// 创建输出文件
	file, err := os.Create(filepath.Join(targetPath, fileName))
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
	}
	defer file.Close()
	
	renderData:=[]string{}
	for _, dbType:= range config.SupportDB{
		renderData=append(renderData, strings.ToUpper(dbType))
	}
	data:=&TableNamingInitMethodData{
		ObjectName: tableData.ObjectName,
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