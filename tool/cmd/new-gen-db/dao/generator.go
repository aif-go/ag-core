package dao

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"ag-core/tool/cmd/new-gen-db/model"
	"ag-core/tool/cmd/new-gen-db/table"
	"ag-core/tool/cmd/new-gen-db/utils"
)

// GenerateDAOFromYAML 从YAML文件生成DAO文件
func GenerateDAOFromYAML(inputFile string, outputDir string, tableName string, moduleName string, dbType string) error {
	// 确保输出目录存在，在用户输入的基础上拼接repository/dao和repository/model
	daoOutputDir := outputDir
	if daoOutputDir != "" {
		// 判断用户输入的是否以/结尾
		if !strings.HasSuffix(daoOutputDir, "/") {
			daoOutputDir += "/"
		}
		daoOutputDir += "repository/dao"
	} else {
		// 如果用户没有指定输出目录，使用默认的repository/dao
		daoOutputDir = "repository/dao"
	}
	modelOutputDir := outputDir
	if modelOutputDir != "" {
		// 判断用户输入的是否以/结尾
		if !strings.HasSuffix(modelOutputDir, "/") {
			modelOutputDir += "/"
		}
		modelOutputDir += "repository/model"
	} else {
		// 如果用户没有指定输出目录，使用默认的repository/model
		modelOutputDir = "repository/model"
	}
	if err := os.MkdirAll(daoOutputDir, 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(modelOutputDir, 0755); err != nil {
		return err
	}

	// 解析YAML文件
	tableDatas, err := YAMLParser(inputFile)
	if err != nil {
		return fmt.Errorf("解析YAML文件失败: %v", err)
	}

	// 获取模块名
	if moduleName == "" {
		moduleName = getModuleName()
	}

	// 解析表名列表，支持多个表名以逗号分隔
	tableNames := utils.ParseCommaSeparatedList(tableName)

	// 生成DAO和模型文件
	for _, tableData := range tableDatas {
		// 如果指定了表名，则只生成指定表的文件
		if len(tableNames) > 0 {
			// 检查当前表名是否在指定的表名列表中
			if !utils.ContainsIgnoreCase(tableNames, tableData.TableName) {
				continue
			}
		}

		// 生成模型文件
		fmt.Printf("正在生成%s的模型文件...\n", tableData.TableName)
		modelPath := filepath.Join(modelOutputDir, strings.ToLower(tableData.TableName)+"_model.go")
		if err := model.GenerateModel(tableData, modelPath); err != nil {
			fmt.Printf("生成模型文件失败: %v，跳过该表\n", err)
			continue
		}
		fmt.Printf("生成模型文件成功: %s\n", tableData.TableName)

		// 生成DAO文件
		fmt.Printf("正在生成%s的DAO文件...\n", tableData.TableName)
		if err := GenerateDAO(tableData, daoOutputDir, moduleName, dbType); err != nil {
			return fmt.Errorf("生成DAO文件失败: %v", err)
		}
		fmt.Printf("生成DAO文件成功: %s\n", tableData.TableName)
	}

	return nil
}

// getModuleName 从go.mod文件中获取模块名
func getModuleName() string {
	// 查找当前目录下的go.mod文件
	goModPath := filepath.Join(".", "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return ""
	}

	// 读取go.mod文件内容
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	// 解析模块名
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimPrefix(line, "module ")
		}
	}

	return ""
}

// GenerateDAO 生成DAO文件
func GenerateDAO(tableData *table.TableData, outputPath string, moduleName string, dbType string) error {
	// 确保输出目录存在
	if err := os.MkdirAll(outputPath, 0755); err != nil {
		return err
	}

	// 生成dao.go文件
	daoFileName := fmt.Sprintf("%s_dao.go", strings.ToLower(tableData.TableName))
	daoPath := fmt.Sprintf("%s/%s", outputPath, daoFileName)
	if err := generateDaoFile(tableData, daoPath, moduleName); err != nil {
		return err
	}

	// 生成tablename_constant.go文件（总是生成）
	constantFileName := fmt.Sprintf("%s_constant.go", strings.ToLower(tableData.TableName))
	constantPath := fmt.Sprintf("%s/%s", outputPath, constantFileName)
	if err := generateConstantFile(tableData, constantPath); err != nil {
		return err
	}

	if len(tableData.SelfQueries) == 0 {
		return nil
	}
	// 生成tablename_namingsql.go文件
	namingSqlFileName := fmt.Sprintf("%s_namingsql.go", strings.ToLower(tableData.TableName))
	namingSqlPath := fmt.Sprintf("%s/%s", outputPath, namingSqlFileName)
	if err := generateNamingSqlFile(tableData, namingSqlPath, dbType); err != nil {
		return err
	}

	// 生成dbtype_tablename_namingsql.go文件
	var dbTypes []string
	if dbType == "" {
		// 不指定时默认生成mysql和db2
		dbTypes = []string{"MYSQL", "DB2"}
	} else {
		// 指定时只生成指定的数据库类型
		dbTypes = []string{strings.ToUpper(dbType)}
	}
	for _, dbType := range dbTypes {
		dbTypeFileName := fmt.Sprintf("%s_%s_namingsql.go", strings.ToLower(dbType), strings.ToLower(tableData.TableName))
		dbTypePath := fmt.Sprintf("%s/%s", outputPath, dbTypeFileName)
		if err := generateDBTypeNamingSqlFile(tableData, dbTypePath, dbType); err != nil {
			return err
		}
	}

	return nil
}

// generateDaoFile 生成dao.go文件
func generateDaoFile(tableData *table.TableData, outputPath string, moduleName string) error {
	// 确保TableData中的ModuleName被设置
	tableData.ModuleName = moduleName

	// 获取DAO模板代码
	code := GetDaoTemplate(tableData)

	// 写入文件
	return os.WriteFile(outputPath, []byte(code), 0644)
}

// generateNamingSqlFile 生成tablename_namingsql.go文件
func generateNamingSqlFile(tableData *table.TableData, outputPath string, dbType string) error {
	// 获取命名SQL模板代码
	code := GetNamingSqlTemplate(tableData, dbType)

	return os.WriteFile(outputPath, []byte(code), 0644)
}

// generateDBTypeNamingSqlFile 生成dbtype_tablename_namingsql.go文件
func generateDBTypeNamingSqlFile(tableData *table.TableData, outputPath string, dbType string) error {
	// 获取数据库类型命名SQL模板代码
	code,err := GetDBTypeNamingSqlTemplate(tableData, dbType)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, []byte(code), 0644)
}

// generateConstantFile 生成tablename_constant.go文件
func generateConstantFile(tableData *table.TableData, outputPath string) error {
	// 获取常量模板代码
	code := GetConstantTemplate(tableData)
	
	// 总是写入文件，因为需要包含命名SQL映射和排除空值字段映射
	return os.WriteFile(outputPath, []byte(code), 0644)
}
