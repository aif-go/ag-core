package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/model"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/other"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/yaml"
)

var (
	inputFile  string
	outputDir  string
	testMode   bool
	tableName  string
	moduleName string
	dbType     string
	keyword    string
	jsonOut    bool
	dryRun     bool
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

	// 创建sheet拆分的子命令
	var sheetCmd = &cobra.Command{
		Use:   "sheet",
		Short: "Split sheet two parts, ddl and custom rule",
		Long:  "Split sheet two parts, ddl and custom rule",
		Run:   runSplitExcelCommand,
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
	dbCmd.Flags().BoolVar(&jsonOut, "json", false, "以 JSON 输出结果（stdout 仅输出 JSON，日志走 stderr）")
	dbCmd.Flags().BoolVar(&dryRun, "dry-run", false, "仅校验并列出将生成的文件，不写盘")

	// 为sheet子命令添加参数
	sheetCmd.Flags().StringVarP(&inputFile, "input", "i", "", "输入yaml文件/目录的路径")
	sheetCmd.Flags().StringVarP(&outputDir, "output", "o", "", "输出拆分后sheet新excel的位置")
	sheetCmd.Flags().StringVarP(&keyword, "keyword", "k", "", "指定拆分sheet的关键字")

	// 将子命令添加到主命令
	rootCmd.AddCommand(yamlCmd)
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(sheetCmd)

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
	// 检查参数是否为空
	if inputFile == "" || outputDir == "" {
		fmt.Println("错误：输入参数不能为空")
		cmd.Usage()
		os.Exit(1)
	}

	// dry-run：生成到临时目录，仅用于校验与枚举产物，不触碰真实输出目录
	runOut := outputDir
	cleanup := func() {}
	if dryRun {
		tmp, err := os.MkdirTemp("", "gendb-dryrun-")
		if err != nil {
			fail(fmt.Errorf("创建临时目录失败: %v", err), jsonOut)
		}
		runOut = tmp
		cleanup = func() { os.RemoveAll(tmp) }
	}

	// 使用dao包中的GenerateDAOFromYAML函数生成DAO文件
	results, err := dao.GenerateDAOFromYAML(inputFile, runOut, tableName, moduleName, dbType)
	cleanup()
	if err != nil {
		fail(err, jsonOut)
	}

	// 成功：按契约输出 JSON 或人类可读摘要
	if jsonOut {
		emitJSONSuccess(results, dryRun)
		return
	}
	emitTextSuccess(results, dryRun)
}

// fail 以既定产出契约终止：JSON 模式先把结构化错误写到 stdout，人类模式只走 stderr；
// 两种模式均以退出码 1 结束（保留 "生成DAO文件失败" 前缀供 e2e 语义校验）。
func fail(err error, asJSON bool) {
	if asJSON {
		emitJSONError(err)
	}
	log.Fatalf("生成DAO文件失败: %v", err)
}

// jsonIssue 类型词汇问题（机读）。
type jsonIssue struct {
	Column string `json:"column"`
	Type   string `json:"type"`
}

// jsonErrorPayload 失败详情（机读）。
type jsonErrorPayload struct {
	Phase   string      `json:"phase"`
	File    string      `json:"file,omitempty"`
	Table   string      `json:"table,omitempty"`
	Issues  []jsonIssue `json:"issues,omitempty"`
	Message string      `json:"message"`
}

// jsonFailure 失败结果（机读）。
type jsonFailure struct {
	OK      bool             `json:"ok"`
	Command string           `json:"command"`
	Error   jsonErrorPayload `json:"error"`
}

// jsonTable 单表产物（机读）。
type jsonTable struct {
	Name  string   `json:"name"`
	Files []string `json:"files"`
}

// jsonSuccess 成功结果（机读）。
type jsonSuccess struct {
	OK      bool        `json:"ok"`
	Command string      `json:"command"`
	DryRun  bool        `json:"dry_run"`
	Tables  []jsonTable `json:"tables"`
}

// emitJSONError 将错误以 JSON 写入 stdout（解析类错误附带结构化字段）。
func emitJSONError(err error) {
	payload := jsonErrorPayload{Phase: "generate", Message: err.Error()}
	var ve *model.ValidationError
	if errors.As(err, &ve) {
		payload.Phase = "parse"
		payload.File = ve.File
		payload.Table = ve.Table
		for _, it := range ve.Issues {
			payload.Issues = append(payload.Issues, jsonIssue{Column: it.Column, Type: it.RawType})
		}
	}
	out, _ := json.Marshal(jsonFailure{OK: false, Command: "db", Error: payload})
	fmt.Println(string(out))
}

// emitJSONSuccess 将成功结果以 JSON 写入 stdout。
func emitJSONSuccess(results []dao.GeneratedTable, dry bool) {
	tables := make([]jsonTable, 0, len(results))
	for _, r := range results {
		tables = append(tables, jsonTable{Name: r.Table, Files: r.Files})
	}
	out, _ := json.Marshal(jsonSuccess{OK: true, Command: "db", DryRun: dry, Tables: tables})
	fmt.Println(string(out))
}

// emitTextSuccess 以人类可读摘要写入 stdout。
func emitTextSuccess(results []dao.GeneratedTable, dry bool) {
	suffix := ""
	if dry {
		suffix = "（dry-run，仅校验未写盘）"
	}
	fmt.Printf("生成完成：%d 个表%s\n", len(results), suffix)
	for _, r := range results {
		fmt.Printf("  - %s\n", r.Table)
		for _, f := range r.Files {
			fmt.Printf("      %s\n", f)
		}
	}
}

// runDbCommand 执行spilt-sheet子命令
func runSplitExcelCommand(cmd *cobra.Command, args []string) {
	fmt.Println("拆分sheet：将原来一个sheet的内容拆分为ddl部分和自定义规则部分两个sheet")

	// 检查参数是否为空
	if inputFile == "" || outputDir == "" {
		fmt.Println("错误：输入参数不能为空")
		cmd.Usage()
		os.Exit(1)
	}

	if keyword == "" {
		keyword = "自定义脚本名字"
	}
	if err := other.SplitExcelByKeyword(inputFile, outputDir, keyword); err != nil {
		log.Fatalf("拆分失败: %v", err)
	}

	fmt.Println("处理完成！")
}
