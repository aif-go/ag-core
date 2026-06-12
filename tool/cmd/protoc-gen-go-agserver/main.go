package main

import (
	"flag"
	"fmt"
	"log/slog"

	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/genserver"
	"github.com/aif-go/ag-core/tool/aggen/parser/protoc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
	model       = flag.String("model", "all", "model to generate, server|client|all default all")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("%s %v\n", pluginName, release)
		return
	}

	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {

		slog.Info(pluginName, "model", *model)

		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		fs := []*protogen.File{}
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			fs = append(fs, f)
		}

		// 解析proto
		geninfo, err := protoc.Parse(gen, fs)
		if err != nil {
			return err
		}

		geninfo.SetPluginName(pluginSortName)
		geninfo.SetBaseVersion(pluginSortName, release)

		// 生成server代码 - 服务接口
		geninfo.Reset()
		err = generator.GenRender(geninfo, genserver.GenServiceTask(*model))
		if err != nil {
			return err
		}

		geninfo.Reset()

		return nil
	})
}
