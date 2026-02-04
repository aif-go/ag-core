package dao

import (
	"os"
	"path/filepath"
	"strings"

	"ag-core/tool/cmd/new-gen-db/model"
	"ag-core/tool/cmd/new-gen-db/table"
)

// YAMLParser 解析YAML文件
func YAMLParser(yamlPath string) ([]*table.TableData, error) {
	var tableDatas []*table.TableData

	// 检查路径是文件还是目录
	info, err := os.Stat(yamlPath)
	if err != nil {
		return nil, err
	}

	var yamlFiles []string
	if info.IsDir() {
		// 遍历目录下的所有yaml文件
		yamlFiles, err = findYAMLFiles(yamlPath)
		if err != nil {
			return nil, err
		}
	} else {
		// 单个yaml文件
		yamlFiles = append(yamlFiles, yamlPath)
	}

	// 解析每个yaml文件
	for _, file := range yamlFiles {
		// 复用common包中的ParseYAML函数
		tableData, err := model.ParseYAML(file, "")
		if err != nil {
			return nil, err
		}
		tableDatas = append(tableDatas, tableData)
	}

	return tableDatas, nil
}

// findYAMLFiles 查找目录下的所有yaml文件
func findYAMLFiles(dir string) ([]string, error) {
	var yamlFiles []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".yaml") {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	return yamlFiles, err
}
