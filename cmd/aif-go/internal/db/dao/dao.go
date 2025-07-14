package dao

import (
	"github.com/spf13/cobra"
	"os/exec"
	"strings"
)

// CmdDao represents the dao command.
var CmdDao = &cobra.Command{
	Use:   "dao",
	Short: "Generate dao",
	Long:  "Generate dao. Example: aif-go db dao",
	Run:   run,
}

var (
	yamlPath       string
	outputPath     string
	moduleName     string
	targetFileName string
)

func init() {
	CmdDao.Flags().StringVarP(&yamlPath, "yaml-path", "i", "internal/idl", "input yaml file path")
	CmdDao.Flags().StringVarP(&outputPath, "output-path", "o", "internal/data/", "output path")
	module, _ := exec.Command("go", "list", "-f", "{{.Module.Path}}", ".").Output()
	CmdDao.Flags().StringVarP(&moduleName, "module-name", "m", strings.Trim(string(module), "\n"), "module name")
	CmdDao.Flags().StringVarP(&targetFileName, "target-file-name", "t", "", "target file name")
}

func run(_ *cobra.Command, args []string) {
	sc := &SchemaConfig{
		ConfigPath:     yamlPath,
		OutputPath:     outputPath,
		ModuleName:     moduleName,
		TargetFileName: targetFileName,
	}
	GenerateSchema(sc)
}
