package model

import (
	"ag-core/tool/cmd/gen-go-db/table"
	"fmt"
	"os"
	"strings"
)

// GenerateModel 生成模型文件
func GenerateModel(data *table.TableData, outputPath string) error {
	// 检查自定义查询的请求参数是否命中索引
	if err := checkQueryArgsIndexes(data); err != nil {
		return err
	}

	// 使用新的模板生成方式
	modelCode := GetModelTemplate(data)

	// 写入文件
	if err := os.WriteFile(outputPath, []byte(modelCode), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// checkQueryArgsIndexes 检查自定义查询的请求参数是否命中索引或主键
func checkQueryArgsIndexes(data *table.TableData) error {
	// 遍历所有自定义查询
	for _, query := range data.SelfQueries {
		// 检查查询的 WhereFields 是否命中索引
		if len(query.WhereFields) == 0 {
			continue
		}

		// 检查是否有任何字段命中索引或主键
		hasHitIndexOrPK := false
		hasHitPrimaryKey := false
		hasHitFirstPrimaryKey := false
		
		// 首先检查是否命中主键（使用PrimaryKeys字段）
		for _, whereField := range query.WhereFields {
			// 检查是否命中任何主键
			for _, pk := range data.PrimaryKeys {
				if strings.EqualFold(pk, whereField) {
					hasHitPrimaryKey = true
					// 检查是否命中第一个主键
					if strings.EqualFold(data.PrimaryKeys[0], whereField) {
						hasHitFirstPrimaryKey = true
					}
					hasHitIndexOrPK = true
					break
				}
			}
			if hasHitIndexOrPK {
				break
			}
		}
		
		// 如果没有命中主键，再检查是否命中索引
		if !hasHitIndexOrPK {
			// 遍历所有索引
			for _, index := range data.Indexes {
				// 检查索引的第一列是否在 WhereFields 中
				if len(index.Columns) > 0 {
					indexFirstCol := index.Columns[0]
					// 检查该列是否在查询的 WhereFields 中
					for _, whereField := range query.WhereFields {
						if strings.EqualFold(indexFirstCol, whereField) {
							hasHitIndexOrPK = true
							break
						}
					}
					if hasHitIndexOrPK {
						break
					}
				}
			}
		}

		// 如果没有命中任何索引或主键，返回错误
		if !hasHitIndexOrPK {
			return fmt.Errorf("查询 %s 的请求参数没有命中任何索引或主键，至少需要包含一个索引的引导列或主键", query.Name)
		}
		
		// 如果有多个主键，并且命中了非第一个主键但没有命中第一个主键，返回错误
		if len(data.PrimaryKeys) > 1 && hasHitPrimaryKey && !hasHitFirstPrimaryKey {
			return fmt.Errorf("查询 %s 命中了非第一个主键但没有命中第一个主键 %s，建议使用第一个主键或确保查询性能", query.Name, data.PrimaryKeys[0])
		}
	}

	return nil
}

