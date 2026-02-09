package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"ag-core/tool/cmd/new-gen-db/dao"
	"ag-core/tool/cmd/new-gen-db/yaml"
)

var (
	inputFile  string
	outputDir  string
	testMode   bool
	tableName  string
	moduleName string
	dbType     string
)

func main() {
	// 创建主命令
	var rootCmd = &cobra.Command{
		Use:   "gendb",
		Short: "Generate database-related files",
		Long:  "A tool to generate database-related files including YAML, model, and DAO files",
	}

	// 创建yaml子命令
	var yamlCmd = &cobra.Command{
		Use:   "yaml",
		Short: "Generate YAML files from Excel",
		Long:  "Generate YAML files from Excel spreadsheet",
		Run:   runYamlCommand,
	}

	// 创建db子命令
	var dbCmd = &cobra.Command{
		Use:   "db",
		Short: "Generate model and DAO files from YAML",
		Long:  "Generate model and DAO files from YAML files",
		Run:   runDbCommand,
	}

	// 为yaml子命令添加参数
	yamlCmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入excel的路径")
	yamlCmd.Flags().StringVarP(&outputDir, "output", "o", "", "最后存放yaml文件的位置")
	yamlCmd.Flags().BoolVarP(&testMode, "test", "t", false, "测试模式，生成示例YAML文件")
	yamlCmd.Flags().StringVarP(&tableName, "table", "T", "", "指定表名，只生成该表的文件")

	// 为db子命令添加参数
	dbCmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入yaml文件/目录的路径")
	dbCmd.Flags().StringVarP(&outputDir, "output", "o", "", "最后存放model和DAO文件的位置")
	dbCmd.Flags().StringVarP(&tableName, "table", "T", "", "指定表名，只生成该表的文件")
	dbCmd.Flags().StringVarP(&moduleName, "module", "m", "", "模块的名字，如果未指定，则查找当前位置的go.mod的值")
	dbCmd.Flags().StringVarP(&dbType, "dbtype", "d", "", "指定数据库类型，可选值：mysql, db2，不指定时默认生成两种")

	// 将子命令添加到主命令
	rootCmd.AddCommand(yamlCmd)
	rootCmd.AddCommand(dbCmd)

	// 执行命令
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// runYamlCommand 执行yaml子命令
func runYamlCommand(cmd *cobra.Command, args []string) {
	// 检查参数是否为空
	if !testMode && (inputFile == "" || outputDir == "") {
		fmt.Println("错误：输入参数不能为空")
		cmd.Usage()
		os.Exit(1)
	}

	// 使用yaml包中的GenerateYAMLFromExcel函数生成YAML文件
	if err := yaml.GenerateYAMLFromExcel(inputFile, outputDir, testMode, tableName); err != nil {
		log.Fatalf("生成YAML文件失败: %v", err)
	}

	fmt.Println("处理完成！")
}

// runDbCommand 执行db子命令
func runDbCommand(cmd *cobra.Command, args []string) {
	fmt.Println("DB生成模式：从YAML文件生成model和DAO代码")

	// 检查参数是否为空
	if inputFile == "" || outputDir == "" {
		fmt.Println("错误：输入参数不能为空")
		cmd.Usage()
		os.Exit(1)
	}

	// 使用dao包中的GenerateDAOFromYAML函数生成DAO文件
	if err := dao.GenerateDAOFromYAML(inputFile, outputDir, tableName, moduleName, dbType); err != nil {
		log.Fatalf("生成DAO文件失败: %v", err)
	}

	fmt.Println("处理完成！")
}
