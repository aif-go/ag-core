package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/emicklei/proto"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// CmdService the service command.
var CmdService = &cobra.Command{
	Use:   "service",
	Short: "Generate the proto service implementations",
	Long:  "Generate the proto service implementations. Example: kratos proto service api/xxx.proto --target-dir=internal/service",
	Run:   run,
}
var targetDir string

func init() {
	CmdService.Flags().StringVarP(&targetDir, "target-dir", "t", "internal/service", "generate target directory")
}

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Please specify the proto file. Example: kratos proto service api/xxx.proto")
		return
	}
	var services []*Service
	for _, f := range args {
		services = generateFile(f)
		generateServiceProxy(services)
	}
	generateService(args)
}

func generateServiceProxy(services []*Service) {
	for _, service := range services {
		iface := &Interface{
			Name:    service.Service,
			PbPkg:   service.Package,
			Methods: service.Methods,
		}
		if _, err := os.Stat(targetDir); os.IsNotExist(err) {
			fmt.Printf("Target directory: %s does not exist\n", targetDir)
			return
		}
		to := filepath.Join(targetDir, strings.ToLower(iface.Name)+"_service_proxy.go")
		if _, err := os.Stat(to); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Already exists: %s\n", to)
			continue
		}
		b, err := iface.execute()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(to, b, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println(to)
	}
}

func generateService(ff []string) {
	pkgs := make([]string, 0)
	annos := make([]Annotate, 0)
	for _, f := range ff {
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
				annos = append(annos, Annotate{Brief: brief, Name: s.Name})
			}),
		)
	}
	uniqPkgs := uniq(pkgs)
	m := &Module{
		Packages: uniqPkgs,
		Names:    annos,
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exist\n", targetDir)
		return
	}
	to := filepath.Join(targetDir, "service.go")
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

func generateFile(f string) []*Service {
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
		res []*Service
	)
	proto.Walk(definition,
		proto.WithOption(func(o *proto.Option) {
			if o.Name == "go_package" {
				pkg = strings.Split(o.Constant.Source, ";")[0]
			}
		}),
		proto.WithService(func(s *proto.Service) {
			cs := &Service{
				Package: pkg,
				Service: serviceName(s.Name),
			}
			for _, e := range s.Elements {
				r, ok := e.(*proto.RPC)
				if !ok {
					continue
				}
				cs.Methods = append(cs.Methods, &Method{
					Service: serviceName(s.Name), Name: serviceName(r.Name), Request: parametersName(r.RequestType),
					Reply: parametersName(r.ReturnsType), Type: getMethodType(r.StreamsRequest, r.StreamsReturns),
				})
			}
			res = append(res, cs)
		}),
	)
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		fmt.Printf("Target directory: %s does not exist\n", targetDir)
		return nil
	}
	for _, s := range res {
		to := filepath.Join(targetDir, strings.ToLower(s.Service)+".go")
		if _, err := os.Stat(to); !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "%s already exists: %s\n", s.Service, to)
			continue
		}
		b, err := s.execute()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(to, b, 0o644); err != nil {
			log.Fatal(err)
		}
		fmt.Println(to)
	}
	return res
}

func getMethodType(streamsRequest, streamsReturns bool) MethodType {
	if !streamsRequest && !streamsReturns {
		return unaryType
	} else if streamsRequest && streamsReturns {
		return twoWayStreamsType
	} else if streamsRequest {
		return requestStreamsType
	} else if streamsReturns {
		return returnsStreamsType
	}
	return unaryType
}

func parametersName(name string) string {
	return strings.ReplaceAll(name, ".", "_")
}

func serviceName(name string) string {
	return toUpperCamelCase(strings.Split(name, ".")[0])
}

func toUpperCamelCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = cases.Title(language.Und, cases.NoLower).String(s)
	return strings.ReplaceAll(s, " ", "")
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
