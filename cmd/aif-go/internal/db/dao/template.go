package dao

import (
	"embed"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/*.tmpl
var TemplateFS embed.FS

func GenerateSchema(schemaConfig *SchemaConfig) {
	// 解析excel文件
	var parse Parse = &YamlParse{}
	//ch := make(chan *TableData, 20)
	//wait := &sync.WaitGroup{}
	// 并行将数据添加到信道
	ch := parse.ParseFile(schemaConfig)
	// 数据渲染template
	for _, schemaData := range ch {
		schemaData.PackageName = schemaConfig.PackageName
		// 要保证main函数没有结束前执行这些内容
		CreteSchemaGoFile(schemaConfig, schemaData)
	}
	//wait.Wait()
}

func CreteSchemaGoFile(conf *SchemaConfig, schemaData *TableData) {

	if conf.OutputPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			log.Panic("获取工作目录失败", err)
		}
		conf.OutputPath = wd + "/repository/"
	}
	log.Println("当前工作目录:", conf.OutputPath)
	createStructGoFile(conf.OutputPath, schemaData)
	createDaoGoFile(conf.OutputPath, schemaData)
}

func createStructGoFile(outpath string, schemaData *TableData) {
	// 定义自定义函数
	funcMap := template.FuncMap{
		"unescaped": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	// 加载模板文件
	tmpl, err := template.New("gorm_struct_yaml.tmpl").Funcs(funcMap).ParseFS(TemplateFS, "templates/gorm_struct_yaml.tmpl")
	if err != nil {
		log.Println("can't find template file", err)
		return
	}

	outpath = outpath + "model/"
	if _, err := os.Stat(outpath); os.IsNotExist(err) {
		error := os.MkdirAll(outpath, 0755)
		if error != nil {
			log.Fatal("创建model目录失败", error)
		}
	}
	// 创建输出文件
	file, err := os.Create(outpath + schemaData.ObjectName + ".go")

	if err != nil {
		log.Println(schemaData.ObjectName, ".go file create error:", err)
		return
	}
	defer file.Close()

	// 渲染模板并写入文件
	err = tmpl.Execute(file, schemaData)
	if err != nil {
		log.Println(schemaData.ObjectName, "render failed,err:", err)
		return
	}
	log.Println("Go file ", schemaData.ObjectName, ".go generated successfully")
}

func createDaoGoFile(outpath string, schemaData *TableData) {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}
	// 加载模板文件
	tmpl, err := template.New("gorm_dao_yaml.tmpl").Funcs(funcMap).ParseFS(TemplateFS, "templates/gorm_dao_yaml.tmpl")

	if err != nil {
		log.Println("can't find template gorm_dao.tmpl file", err)
		return
	}

	outpath = outpath + "dao/"
	if _, err := os.Stat(outpath); os.IsNotExist(err) {
		error := os.MkdirAll(outpath, 0755)
		if error != nil {
			log.Fatal("创建dao目录失败", error)
		}
	}
	// 创建输出文件
	daoFileName := schemaData.ObjectName + "Dao" + ".go"
	outpath = filepath.Join(outpath, daoFileName)
	file, err := os.Create(outpath)

	if err != nil {
		log.Println(schemaData.ObjectName, ".go file create error:", err)
		return
	}
	defer file.Close()
	// 渲染模板并写入文件
	err = tmpl.Execute(file, schemaData)
	if err != nil {
		log.Println(daoFileName, "render failed,err:", err)
		return
	}
	log.Println("Go file ", daoFileName, " generated successfully")
}
