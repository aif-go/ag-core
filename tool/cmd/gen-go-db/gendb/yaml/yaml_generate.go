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

	listyaml, err:= findYamlFiles(config.DbTemplatePath,false)
	if err!=nil{
		fmt.Printf("获取路径:%v下的yaml文件列表失败：%v\n", config.OutputPath,err)
		return nil,err
	}
	yamlDataList:=make([]*YamlDataConfig,0)
	for _, yamlfilepath:=range listyaml{
		// 1. 读取YAML文件
		yamlFile, err := os.ReadFile(yamlfilepath)
		if err != nil {
			fmt.Printf("读取YAML文件失败：%v\n", err)
			continue
		}
		// 2. 解析YAML到结构体
		var yamlFileData *YamlDataConfig=&YamlDataConfig{}
		err = yaml.Unmarshal(yamlFile, yamlFileData)
		if err != nil {
			fmt.Printf("解析YAML失败：%v\n", err)
			continue
		}
		// 设置数据库类型
		yamlFileData.DatabaseTable.DbType=config.DbType
		yamlDataList = append(yamlDataList, yamlFileData)
	}
	
	return 	&YamlAllTableData{yamlDataList}, nil 
}

// Generate 构建dao,entity,sql文件的方法
func (generate *YamlGenerate) Generate(config *render.AGInfraStructrueConfig, yamlAllTableData *YamlAllTableData) error {
	dataConvert := YamlDataConvert{}
	tableDataList := dataConvert.Convert(yamlAllTableData.YamlDataList)
	supportTableSize:=len(config.SupportTables)
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
				if err := render.RenderDaoTemplate(config.DaoPath, data); err != nil {
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


// FindYamlFiles 查找指定目录下所有.yaml/.yml文件（支持递归子目录）
// 参数：
//   rootDir: 根目录路径
//   recursive: 是否递归遍历子目录
// 返回：
//   []string: 符合条件的文件路径列表
//   error: 错误信息
func findYamlFiles(rootDir string, recursive bool) ([]string, error) {
	var yamlFiles []string

	// 遍历目录
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		// 处理遍历过程中的错误（如权限不足）
		if err != nil {
			return fmt.Errorf("访问路径 %s 失败：%v", path, err)
		}

		// 跳过目录（仅处理文件）
		if info.IsDir() {
			// 如果不递归，跳过当前目录以外的子目录
			if !recursive && path != rootDir {
				return filepath.SkipDir
			}
			return nil
		}

		// 过滤.yaml/.yml后缀的文件（忽略大小写）
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			yamlFiles = append(yamlFiles, path)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败：%v", err)
	}

	return yamlFiles, nil
}