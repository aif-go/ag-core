package main

import (
	"ag-core/cmd/protoc-gen-go-agkitex/aggen"
	"ag-core/cmd/protoc-gen-go-agkitex/genner"
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const (
	pluginName     = "protoc-gen-go-agkitex"
	pluginSortName = "agkitex"
)

var (
	showVersion = flag.Bool("version", false, "print the version and exit")
	// omitempty       = flag.Bool("omitempty", true, "omit if google.api is empty")
	// omitemptyPrefix = flag.String("omitempty_prefix", "", "omit if google.api is empty")
)

func main() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-agkitex %v\n", release)
		return
	}
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, f := range gen.Files {
			if !f.Generate {
				continue
			}

			// 解析提取PackageInfo
			pi := aggen.ParsePackageInfo_protoc(gen, f)
			pi.Version = release
			pi.PluginName = pluginName

			/* ========kitex 代码生成 逻辑========= */
			// kitex: g.ProtoGen(f, w) => hello.pb.go
			// ProtoGen 应该大部分被protoc_go代替了
			// f.Enums
			//  └  EnumGen()
			// f.Messages
			//  ├ m.Enums
			//  │ └ EnumGen()
			//  └ MessageGen()

			// kitex: genKitexServiceInterface(f, w, c.StreamX) => hello.pb.go TODO

			// kitex: pg.generateClientServerFiles(f, p) => client.go、server.go、hello.go TODO
			/* ================= */

			err := aggen.GenService(pi, genner.Genner)
			// 生成client、server、service.go
			// err := genner.GenerateClientServerFiles(f, pi)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
