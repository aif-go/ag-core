package main

import (
	"flag"
	"fmt"
	"log/slog"

	"ag-core/tool/aggen/genagservice"
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/genhertz"
	"ag-core/tool/aggen/genkitex"
	"ag-core/tool/aggen/genserver"
	"ag-core/tool/aggen/parser/protoc"

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

		// slog.Info("PwdGoModPath", "X", geninfo.GlobalInfo.ModuleInfo.PwdGoModPath)
		// 调用代码生成，材料：import生成函数、task创建函数，还有区分阶段

		// 生成server代码 - 服务接口
		geninfo.Reset()
		err = generator.GenRender(geninfo, genserver.GenServiceTask(*model))
		if err != nil {
			return err
		}

		// 生成kitex代码
		geninfo.Reset()
		geninfo.ResetVersion()
		geninfo.SetVersion("kitex", "v0.14.1") // TODO kitex的版本怎么获取
		err = generator.GenRender(geninfo, genkitex.KitexGenServiceTask(*model))
		if err != nil {
			return err
		}

		// 生成hertz代码
		geninfo.Reset()
		geninfo.ResetVersion()
		geninfo.SetVersion("hertz", "v0.10.0") // TODO hertz的版本怎么获取
		err = generator.GenRender(geninfo, genhertz.HertzGenServiceTask(*model))
		if err != nil {
			return err
		}

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
