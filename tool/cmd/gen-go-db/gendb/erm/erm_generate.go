package erm

import (
	"encoding/xml"
	"os"
	"strings"
	"sync"

	"ag-core/tool/cmd/gen-go-db/gendb/render"
)

type ErmGenerate struct {
}

// ParseTemplateFile 解析模板文件
func (generate *ErmGenerate) ParseTemplateFile(config *render.AGInfraStructrueConfig) (*Diagram, error) {

	data, err := os.ReadFile(config.DbTemplatePath)
	if err != nil {
		return nil, err
	}

	diagram := &Diagram{}
	if err := xml.Unmarshal(data, diagram); err != nil {
		return nil, err
	}
	diagram.DbType = strings.ToUpper(config.DbType)
	return diagram, nil
}

func (generate *ErmGenerate) Generate(config *render.AGInfraStructrueConfig, diagram *Diagram) error {
	dataConvert := ErmDataConvert{}
	tableDataList := dataConvert.Convert(diagram)
	supportTableSize := len(config.SupportTables)

	// 过滤需要生成的表
	filteredTables := make([]*render.TableData, 0, len(tableDataList))
	for _, tableData := range tableDataList {
		if _, ok := config.SupportTables[tableData.TableName]; !ok && supportTableSize != 0 {
			continue
		}
		if len(tableData.NamingSqlMap) != 0 {
			tableData.NamingSqlMapEnable = true
		}
		tableData.ModelName = config.PackageNamePrefix
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
