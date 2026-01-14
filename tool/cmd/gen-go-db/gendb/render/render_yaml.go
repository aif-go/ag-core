package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// RenderYaml 构建yaml文件
func RenderYaml(config *AGInfraStructrueConfig, tableData *TableData) error {
	yamlData := &YamlData{
		ModelName:        tableData.ModelName,
		TableName:        tableData.TableName,
		ColumnList:       tableData.ColumnList,
		GeneralIndexList: tableData.GeneralIndexList,
		UniqueIndexList:  tableData.UniqueIndexList,
		PrimaryKeyList:   strings.Split(tableData.PrimaryKeyList, ","),
		RNamingSqlList:   tableData.RNamingSqlList,
		CUDNamingSqlList: tableData.CUDNamingSqlList,
		Encode:           tableData.Encode,
		Engine:           tableData.Engine,
		Sort:             tableData.Sort,
	}
	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("解析yaml对象数据失败:%s: %w", yamlData.TableName, err)
	}

	filepath := filepath.Join(config.YamlPath, yamlData.TableName+".yaml")
	err = os.WriteFile(filepath, yamlBytes, 0644)
	if err != nil {
		return fmt.Errorf("构建%s的yaml文件失败: %w", yamlData.TableName, err)
	}
	// 修改文件权限为只读权限
	// os.Chmod(filepath, 0444)
	return nil
}
