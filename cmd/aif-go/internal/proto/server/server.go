package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
)

// CmdServer the server command.
var CmdServer = &cobra.Command{
	Use:   "server",
	Short: "Generate the fx server module definition",
	Long:  "Generate the fx server module definition. Example: kratos proto server api/xxx.proto --target-dir=internal/server",
	Run:   run,
}
var targetDir string
var targetFile string

func init() {
	CmdServer.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/server", "generate target directory")
	CmdServer.Flags().StringVarP(&targetFile, "target-file", "n", "server", "generate target file")
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto server api/xxx.proto")
		return
	}
	generateServer(args)
}

func generateServer(f []string) {
	pkgs := make([]string, 0)
	names := make([]string, 0)
	for _, f := range f {
		reader, err := os.Open(f)
		if err != nil {
			log.Fatal(err)
		}
		defer reader.Close()

		parser := proto.NewParser(reader)
		definition, err := parser.Parse()
		if err != nil {
			log.Fatal(err)
		}

		var (
			pkg string
		)

		proto.Walk(definition,
			proto.WithOption(func(o *proto.Option) {
				if o.Name == "go_package" {
					pkg = strings.Split(o.Constant.Source, ";")[0]
					pkgs = append(pkgs, pkg)
				}
			}),
			proto.WithService(func(s *proto.Service) {
				n := strings.LastIndex(pkg, "/")
				brief := pkg[n+1:]
				names = append(names, fmt.Sprintf("%s.Fx%sHTTPModule", brief, s.Name))
				names = append(names, fmt.Sprintf("%s.Fx%sGRPCModule", brief, s.Name))
			}),
		)
	}
	uniqPkgs := uniq(pkgs)
	m := &Module{
		Packages: uniqPkgs,
		Names:    names,
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exist\n", targetDir)
		return
	}
	to := filepath.Join(targetDir, targetFile+".go")
	if _, err := os.Stat(to); !os.IsNotExist(err) {
		backupPath := generateBackupName(to)
		fmt.Fprintf(os.Stderr, "Already exists: %s,backup to: %s\n", to, backupPath)
		if err := os.Rename(to, backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error backing up file: %v\n", err)
			return
		}
	}
	b, err := m.execute()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile(to, b, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Println(to)
}

func uniq(input []string) []string {
	unique := make(map[string]bool)
	for _, v := range input {
		unique[v] = true
	}
	// 将map中的键提取到切片中
	result := make([]string, 0, len(unique))
	for k := range unique {
		result = append(result, k)
	}
	return result
}

func generateBackupName(originalPath string) string {
	// 获取当前时间并格式化为 YYYY-MM-DD-HHMM
	timestamp := time.Now().Format("2006-01-02-1504")

	// 获取文件目录和基本文件名
	dir, file := filepath.Split(originalPath)

	// 组合新文件名：原文件名 + 时间戳 + 扩展名
	backupFile := fmt.Sprintf("%s.%s", file, timestamp)

	return filepath.Join(dir, backupFile)
}
