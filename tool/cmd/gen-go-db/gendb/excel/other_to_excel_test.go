package excel

import (
	"testing"

	"ag-core/tool/cmd/gen-go-db/gendb/erm"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
)

func TestOtherToExcel(t *testing.T) {

	generate:= &erm.ErmGenerate{

	}
	config:=&render.AGInfraStructrueConfig{
		BaseConfig: render.BaseConfig{
			DbTemplatePath: "C:/Users/songbing/Desktop/goerm/mps.erm",
			// DbTemplatePath: "C:/Users/songbing/Desktop/goerm/mps.xlsx",
			PackageNamePrefix: "ag-core/tool/cmd/gen-go-db/gendb",
			OutputPath: "./",
			DbType: "mysql",
		},
		GenerateOptions: render.GenerateOptions{
			Entityable: true,
			Sqlable: true,
			Daoable: true,
		},
		SupportConfig: render.SupportConfig{
			SupportDB: []string{"mysql","db2"},
		},
	} 

	diagram, err := generate.ParseTemplateFile(config)
	if err != nil {
		t.Fatalf("解析模板文件失败: %v", err)
	}
	convert:=&erm.ErmDataConvert{
	}

	list:=convert.Convert(diagram)
	OtherToExcel(config, list)
}