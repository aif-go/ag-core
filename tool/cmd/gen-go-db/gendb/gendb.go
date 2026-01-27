package gendb

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"ag-core/tool/cmd/gen-go-db/gendb/erm"
	"ag-core/tool/cmd/gen-go-db/gendb/excel"
	"ag-core/tool/cmd/gen-go-db/gendb/generate"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"ag-core/tool/cmd/gen-go-db/gendb/yaml"
)

func GenerateDBGoFile(config *render.AGInfraStructrueConfig) error {

	// 校验参数
	err := verifyConfig(config)
	if err != nil {
		return err
	}
	if config.OutputPath == "" {
		return nil
	}
	// 组装路径
	if config.Daoable {
		config.DaoPath = filepath.Join(config.OutputPath, render.DaoPathSuffix)
		if err := mkdir(config.DaoPath); err != nil {
			return err
		}
		// config.YamlPath = filepath.Join(config.OutputPath, render.YamlPathSuffix)
		// if err := mkdir(config.YamlPath); err != nil {
		// 	return err
		// }
	}
	if config.Entityable {
		config.EntityPath = filepath.Join(config.OutputPath, render.EntityPathSuffix)
		if err := mkdir(config.EntityPath); err != nil {
			return err
		}
	}
	if config.Sqlable {
		config.SqlPath = filepath.Join(config.OutputPath, render.SqlPathSuffix)
		if err := mkdir(config.SqlPath); err != nil {
			return err
		}
	}
	// 截取DB文件数据的后缀名
	// filenamesuffix := filepath.Ext(config.DbTemplatePath)

	// var generate generate.Generate[any]
	// switch filenamesuffix {
	// case ".erm":
	// 	erm := &erm.ErmGenerate{}
	// 	if err := execute(config, erm); err != nil {
	// 		return err
	// 	}
	// case ".xlsx":
	// 	xlsx := &excel.ExcelGenerate{}
	// 	if err := execute(config, xlsx); err != nil {
	// 		return err
	// 	}
	// case ".yaml", ".yml":
		yaml := &yaml.YamlGenerate{}
		if err := execute(config, yaml); err != nil {
			return err
		}

	// default:
	// 	return errors.New("当前仅支持erm,excel,yaml文件")
	// }
	log.Println("构建go文件执行完毕")
	return nil
}

// GenerateYamlFile 生成yaml文件
func GenerateYamlFile(config *render.AGInfraStructrueConfig) error {
	if config.OutputPath == "" {
		return errors.New("输出路径未配置")
	}
	config.OutputPath = config.OutputPath + render.YamlPathSuffix
	mkdir(config.OutputPath)
	err := yaml.GenerateYamlFromExcel(config)
	if err != nil {
		return err
	}
	return nil
}

// GenerateExcelFile 转换其余文档为excel文件
func GenerateExcelFile(config *render.AGInfraStructrueConfig) error {
	if config.OutputPath == "" {
		return nil
	}
	generate := &erm.ErmGenerate{}
	diagram, err := generate.ParseTemplateFile(config)
	if err != nil {
		return err
	}
	convert := &erm.ErmDataConvert{}
	list := convert.Convert(diagram)
	excel.OtherToExcel(config, list)
	// if err != nil {
	// 	return err
	// }
	// 	return err
	// }
	log.Println("构建go文件执行完毕")
	return nil
}

func execute[T any, R generate.Generate[T]](config *render.AGInfraStructrueConfig, gen R) error {
	data, err := gen.ParseTemplateFile(config)
	if err != nil {
		return err
	}
	if err := gen.Generate(config, data); err != nil {
		return err
	}
	return nil
}

func verifyConfig(config *render.AGInfraStructrueConfig) error {
	if config.Entityable || config.Daoable || config.Sqlable {
		if config.OutputPath == "" {
			return errors.New("输出路径不能为空")
		}
		if config.DbTemplatePath == "" {
			return errors.New("db源数据不能为空")
		}
		if config.PackageNamePrefix == "" {
			return errors.New("go模块的前缀不能为空")
		}
		// if config.DbType == "" {
		// 	return errors.New("数据库的类型不能为空")
		// }
	}
	return nil
}

func mkdir(dir string) error {
	// 构建目录
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}
	return nil
}
