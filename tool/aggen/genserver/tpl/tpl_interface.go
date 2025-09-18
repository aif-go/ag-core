package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
)

// === PackageGroup级别 ===
// InterfaceImportsSetter 设置Import部分信息
var InterfaceImportsSetter = func(geni *types.GennerInfo) error {
	geni.AddImports("context")
	pkgg := geni.PackageGroup
	for _, pkg := range pkgg.PackageInfos {
		for _, s := range pkg.Services {
			for _, m := range s.Methods {
				if m.IsStreaming && pkg.Codec == "protobuf" {
					geni.AddImport("streaming", "github.com/cloudwego/kitex/pkg/streaming")
				}
			}
		}
	}
	return nil
}

// InterfaceTpl is the template for generating server.go.
var InterfaceTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	interfaceTpl +
	""

var interfaceTpl string = `
{{- $LMI := .ModuleInfo}}     {{/* 模块信息 */}}
{{- $LGI := .GlobalInfo}}     {{/* 全局信息 */}}
{{- $LPkgInfo := .PkgInfo}}   {{/* 包信息 */}}
{{- $LPG := .PackageGroup}}   {{/* 包组信息 */}}

{{- range $LPI := $LPG.PackageInfos}}
	{{- range $LPI.Services}}
		{{- $LSI := .}}
		{{- $LSName := $LSI.ServiceName}}

		// {{$LSName}} defines the server interface for the {{$LSName}} service.
		type {{$LSName}} interface { 
			{{- range $LSI.Methods}}
				{{- $LMethod := .}}
			
				{{- $LArg := index $LMethod.Args 0 }} {{/* 规定接口参数只有一个参数 */}}
				{{- $LResp := $LMethod.Resp}}
				{{- if and $LMethod.ClientStreaming $LMethod.ServerStreaming }}
					{{$LMethod.Name}}(stream {{$LSName}}_{{$LMethod.Name}}Server) (err error)
				{{- else if $LMethod.ClientStreaming}}
					{{$LMethod.Name}}(stream {{$LSName}}_{{$LMethod.Name}}Server) (err error)
				{{- else if $LMethod.ServerStreaming}}
					{{$LMethod.Name}}(req *{{$LArg.RawType}}, stream {{$LSName}}_{{$LMethod.Name}}Server) (err error)
				{{- else}}
					{{$LMethod.Name}}(context.Context, *{{$LArg.RawType}}) (*{{$LResp.RawType}}, error)
				{{- end}}				

			{{- end}}
		}

		{{- range $LSI.Methods}}
			{{- $LMethod := .}}

			{{- $LArg := index $LMethod.Args 0 }} {{/* 规定接口参数只有一个参数 */}}
			{{- $LResp := $LMethod.Resp}}
			{{- if and $LMethod.ClientStreaming $LMethod.ServerStreaming }}
				// {{$LSName}}_{{$LMethod.Name}}Server
				type {{$LSName}}_{{$LMethod.Name}}Server interface {
					streaming.Stream
					Recv() (*{{$LArg.RawType}}, error)
					Send(*{{$LResp.RawType}}) error
				}
			{{- else if $LMethod.ClientStreaming}}
				// {{$LSName}}_{{$LMethod.Name}}Server
				type {{$LSName}}_{{$LMethod.Name}}Server interface {
					streaming.Stream
		        	Recv() (*{{$LArg.RawType}}, error)
		        	SendAndClose(*{{$LResp.RawType}}) error	
				}
			{{- else if $LMethod.ServerStreaming}}
				// {{$LSName}}_{{$LMethod.Name}}Server
				type {{$LSName}}_{{$LMethod.Name}}Server interface {
					streaming.Stream
					Send(*{{$LResp.RawType}}) error
				}
			{{- else}}
				{{/* 没事做 */}}
			{{- end}}
		{{- end}}


	{{- end}}
{{- end}}

`
