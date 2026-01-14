package erm

import (
	"log"
	"testing"

	"ag-core/tool/cmd/gen-go-db/gendb/render"
)

func TestERMParse(t *testing.T) {

	config := &render.AGInfraStructrueConfig{
		BaseConfig: render.BaseConfig{
			DbTemplatePath:    "C:/Users/songbing/Desktop/goerm/mps.erm",
			PackageNamePrefix: "ag-core/tool/cmd/gen-go-db/gendb",
			DbType:            "db2",
		},
		GenerateOptions: render.GenerateOptions{
			Sqlable: true,
			// Daoable: true,
			// Entityable: true,
		},
		PathConfig: render.PathConfig{
			// OutputPath: "C:/Users/songbing/Desktop/goerm/",
			DaoPath:    "C:/Users/songbing/Desktop/goerm/dao/",
			EntityPath: "C:/Users/songbing/Desktop/goerm/entity/",
			SqlPath:    "C:/Users/songbing/Desktop/goerm/sql/",
		},
	}

	// var generate Generate
	generate := &ErmGenerate{}
	diagram, err := generate.ParseTemplateFile(config)
	if err != nil {
		t.Fatalf("解析模板文件失败: %v", err)
	}
	err = generate.Generate(config, diagram)
	if err != nil {
		t.Fatalf("生成文件失败: %v", err)
	}
	log.Println("执行结果")
}
