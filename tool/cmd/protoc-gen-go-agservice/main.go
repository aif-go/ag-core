package main

import (
	"flag"
	"fmt"

	"ag-core/tool/aggen/genagservice"
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/parser/protoc"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
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

		// 生成 agservice 模板代码
		geninfo.Reset()
		geninfo.ResetVersion()
		err = generator.GenRender(geninfo, genagservice.GenServiceTask())
		if err != nil {
			return err
		}

		geninfo.Reset()

		return nil
	})
}
