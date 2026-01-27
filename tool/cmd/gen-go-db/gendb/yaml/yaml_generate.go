package yaml

import (
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

type YamlGenerate struct {
}

// ParseTemplateFile 解析模板文件
func (generate *YamlGenerate) ParseTemplateFile(config *render.AGInfraStructrueConfig) (*YamlAllTableData, error) {

	listyaml, err := findYamlFiles(config.DbTemplatePath, false)
	if err != nil {
		fmt.Printf("获取路径:%v下的yaml文件列表失败：%v\n", config.OutputPath, err)
		return nil, err
	}
	yamlDataList := make([]*YamlDataConfig, 0)
	for _, yamlfilepath := range listyaml {
		// 1. 读取YAML文件
		yamlFile, err := os.ReadFile(yamlfilepath)
		if err != nil {
			fmt.Printf("读取YAML文件失败：%v\n", err)
			continue
		}
		// 2. 解析YAML到结构体
		var yamlFileData *YamlDataConfig = &YamlDataConfig{}
		err = yaml.Unmarshal(yamlFile, yamlFileData)
		if err != nil {
			fmt.Printf("解析YAML失败：%v\n", err)
			continue
		}
		// 设置数据库类型
		yamlFileData.DatabaseTable.DbType = config.DbType
		yamlDataList = append(yamlDataList, yamlFileData)
	}

	return &YamlAllTableData{yamlDataList}, nil
}

// Generate 构建dao,entity,sql文件的方法
func (generate *YamlGenerate) Generate(config *render.AGInfraStructrueConfig, yamlAllTableData *YamlAllTableData) error {
	dataConvert := YamlDataConvert{}
	tableDataList := dataConvert.Convert(yamlAllTableData.YamlDataList)
	supportTableSize := len(config.SupportTables)
	// 过滤需要生成的表
	filteredTables := make([]*render.TableData, 0, len(tableDataList))
	for _, tableData := range tableDataList {
		if _, ok := config.SupportTables[tableData.TableName]; !ok && supportTableSize != 0 {
			continue
		}
		tableData.ModelName = config.PackageNamePrefix
		tableData.DbType = strings.ToUpper(config.DbType)
		if len(tableData.NamingSqlMap) != 0 {
			tableData.NamingSqlMapEnable = true
		}
		filteredTables = append(filteredTables, tableData)
	}

	// 使用并发方式生成文件
	var wg sync.WaitGroup
	errCh := make(chan error, len(filteredTables))

	for _, tableData := range filteredTables {
		wg.Add(1)
		go func(data *render.TableData) {
			defer wg.Done()

			// 构建yaml文件
			// if err := render.RenderYaml(config, data); err != nil {
			// 	errCh <- err
			// 	return
			// }

			// 构建struct.go文件
			if config.Entityable {
				if err := render.RenderEnityTemplate(config.EntityPath, data); err != nil {
					errCh <- err
					return
				}
			}

			// 构建dao.go文件
			if config.Daoable {
				// if err := render.RenderDaoTemplate(config.DaoPath, data); err != nil {
				if err := render.RenderDaoTemplate_hzw(config.DaoPath, data); err != nil {
					errCh <- err
					return
				}
				if data.NamingSqlMapEnable {
					if err := render.RenderNamingSqlConstant(config, data); err != nil {
						errCh <- err
						return
					}
					if err := render.RenderTableNamingSql(config, data); err != nil {
						errCh <- err
						return
					}
				}
			}

			// 构建sql文件
			if config.Sqlable {
				if err := render.RenerDDLTemplate(config, data); err != nil {
					errCh <- err
					return
				}
			}
		}(tableData)
	}

	// 等待所有goroutine完成
	go func() {
		wg.Wait()
		close(errCh)
	}()

	// 检查是否有错误发生
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// findYamlFiles 查找指定目录下所有.yaml/.yml文件（支持递归子目录）
// 参数：
//
//	rootDir: 根目录路径
//	recursive: 是否递归遍历子目录
//
// 返回：
//
//	[]string: 符合条件的文件路径列表
//	error: 错误信息
func findYamlFiles(rootDir string, recursive bool) ([]string, error) {
	var yamlFiles []string

	// 检查 rootDir 是否为文件
	info, err := os.Stat(rootDir)
	if err != nil {
		return nil, fmt.Errorf("访问路径失败: %w", err)
	}

	// 如果是文件，检查扩展名
	if !info.IsDir() {
		ext := strings.ToLower(filepath.Ext(rootDir))
		if ext == ".yaml" || ext == ".yml" {
			return []string{rootDir}, nil
		}
		// 如果不是 YAML 文件，返回空列表
		return []string{}, nil
	}

	// 如果是目录，遍历查找 YAML 文件
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// 如果不递归且不是根目录，则跳过子目录
			if !recursive && path != rootDir {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名是否为 .yaml 或 .yml
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			yamlFiles = append(yamlFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %w", err)
	}

	return yamlFiles, nil
}
