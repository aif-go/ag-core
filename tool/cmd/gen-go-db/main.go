package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"ag-core/tool/cmd/gen-go-db/gendb"
	"ag-core/tool/cmd/gen-go-db/gendb/render"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var (
	version = "dev"     // 版本号
	commit  = "unknow"  // git提交的哈希码
	date    = "unknown" // 构建时间
)

func main() {

	rootCmd := &cobra.Command{
		Use:   "agdb",
		Short: "agcore db模块的命令",
		Run: func(cmd *cobra.Command, args []string) {
			// 业务逻辑...
			fmt.Println("ag-core的db命令...")
		},
	}
	// 添加子命令
	rootCmd.AddCommand(VersioCommand())
	rootCmd.AddCommand(DbCommand())
	rootCmd.AddCommand(ExcelCommand())
	rootCmd.AddCommand(YamlCommand())
	// 编译时注入版本（同方案一的 ldflags 方式）
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// VersioCommand 展示版本号
func VersioCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "显示版本信息",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("版本: %s\n", version)
			fmt.Printf("提交: %s\n", commit)
			fmt.Printf("构建时间: %s\n", date)
		},
	}
}

// DbCommand 构建系统的db模块
func DbCommand() *cobra.Command {
	var agconfig = &render.AGInfraStructrueConfig{}
	dbommand := &cobra.Command{
		Use:   "db",
		Short: "自动化生成dao、entity、ddl、yaml文件",
		Run: func(cmd *cobra.Command, args []string) {
			dbtype, _ := cmd.Flags().GetString("dbtype")
			outpath, _ := cmd.Flags().GetString("opath")
			packagename, _ := cmd.Flags().GetString("packagename")
			inpath, _ := cmd.Flags().GetString("ipath")
			tablenames, _ := cmd.Flags().GetString("tablenames")
			fmt.Println("命令执行开始....")
			if packagename == "" {
				moduleByte, err := exec.Command("go", "list", "-f", "{{.Module.Path}}", ".").Output()
				if err != nil {
					fmt.Println("获取模块路径失败:", err)
					return
				}
				packagename = strings.TrimSpace(string(moduleByte))
				fmt.Println(packagename)
			}
			agconfig.PackageNamePrefix = packagename
			agconfig.OutputPath = outpath
			agconfig.DbTemplatePath = inpath
			agconfig.SupportDB = []string{"mysql", "db2"}
			agconfig.DbType = dbtype
			var tableMap map[string]string
			if tablenames != "" {
				tableMap = lo.SliceToMap(strings.Split(strings.ToLower(tablenames), ","), func(s string) (string, string) { return strings.ToUpper(s), s })
			}
			agconfig.SupportTables = tableMap
			agconfig.Entityable = true
			agconfig.Daoable = true
			agconfig.Sqlable = true
			err := gendb.GenerateDBGoFile(agconfig)
			if err != nil {
				fmt.Println("生成失败:", err)
			}
		},
	}
	// 以下动作帮助--help命令的时候展示出对应的flags
	dbommand.Flags().StringP("dbtype", "d", "db2", "db类型,当前仅支持db2和mysql")
	dbommand.Flags().StringP("opath", "o", "./", "go文件生成到项目位置,默认是当前执行命令的位置")
	dbommand.Flags().StringP("packagename", "p", "", "go文件中引入包的路径前缀,默认当前执行命令的路径")
	dbommand.Flags().StringP("ipath", "i", "", "数据库模块的文档位置")
	dbommand.Flags().StringP("tablenames", "t", "", "指定模板中对应的表名,默认模板中全部的表名")
	return dbommand
}

func ExcelCommand() *cobra.Command {
	var agconfig = &render.AGInfraStructrueConfig{}
	excelCommand := &cobra.Command{
		Use:   "excel",
		Short: "将erm等其他文件自动转换为excel文件工具",
		Run: func(cmd *cobra.Command, args []string) {
			outpath, _ := cmd.Flags().GetString("opath")
			inpath, _ := cmd.Flags().GetString("ipath")
			tablenames, _ := cmd.Flags().GetString("tablenames")
			agconfig.OutputPath = outpath
			agconfig.DbTemplatePath = inpath
			var tableMap map[string]string
			if tablenames != "" {
				tableMap = lo.SliceToMap(strings.Split(strings.ToLower(tablenames), ","), func(s string) (string, string) { return strings.ToUpper(s), s })
			}
			agconfig.SupportTables = tableMap
			gendb.GenerateExcelFile(agconfig)
		},
	}
	// 以下动作帮助--help命令的时候展示出对应的flags
	excelCommand.Flags().StringP("opath", "o", "./", "生成的excel文件存放位置")
	excelCommand.Flags().StringP("ipath", "i", "", "目标文件位置")
	excelCommand.Flags().StringP("tablenames", "t", "", "指定模板中对应的表名,默认模板中全部的表名")
	return excelCommand
}

// YamlCommand 生成yaml idl
func YamlCommand() *cobra.Command {
	var agconfig = &render.AGInfraStructrueConfig{}
	excelCommand := &cobra.Command{
		Use:   "yaml",
		Short: "将erm等其他文件自动转换为excel文件工具",
		Run: func(cmd *cobra.Command, args []string) {
			outpath, _ := cmd.Flags().GetString("opath")
			inpath, _ := cmd.Flags().GetString("ipath")
			tablenames, _ := cmd.Flags().GetString("tablenames")
			agconfig.OutputPath = outpath
			agconfig.DbTemplatePath = inpath
			var tableMap map[string]string
			if tablenames != "" {
				tableMap = lo.SliceToMap(strings.Split(strings.ToLower(tablenames), ","), func(s string) (string, string) { return strings.ToUpper(s), s })
			}
			agconfig.SupportTables = tableMap
			gendb.GenerateYamlFile(agconfig)
		},
	}
	// 以下动作帮助--help命令的时候展示出对应的flags
	excelCommand.Flags().StringP("opath", "o", "./", "生成的yaml文件存放位置")
	excelCommand.Flags().StringP("ipath", "i", "", "目标文件位置")
	excelCommand.Flags().StringP("tablenames", "t", "", "指定模板中对应的表名,默认模板中全部的表名")
	return excelCommand
}
