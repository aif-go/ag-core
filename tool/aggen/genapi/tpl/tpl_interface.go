package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
)

// === PackageGroup级别 ===
// InterfaceIsSkip 判断是否跳过该任务
var InterfaceIsSkip = func(geni *types.GennerInfo) bool {
	pkgg := geni.PackageGroup

	for _, pkg := range pkgg.PackageInfos {
		if len(pkg.Services) > 0 {
			return false
		}
	}

	return true
}

// InterfaceImportsSetter 设置Import部分信息
var InterfaceImportsSetter = func(geni *types.GennerInfo) error {
	pkgg := geni.PackageGroup
	pkginfo := pkgg.PkgInfo

	for _, pkg := range pkgg.PackageInfos {
		for _, s := range pkg.Services {
			for _, m := range s.Methods {
				if m.IsStreaming && pkg.Codec == "protobuf" {
					geni.AddImport("streaming", "github.com/cloudwego/kitex/pkg/streaming")
				}

				if !m.IsStreaming {
					geni.AddImports("context") // 普通方法入参包含context
				}

				// 入参
				for _, arg := range m.Args {
					for _, dep := range arg.Deps {
						if dep.ImportPath != pkginfo.ImportPath { // 同包下无需import，否则循环依赖
							// TODO  非同包、目标文件又是本地生成的，要修改依赖为当前model
							geni.AddImport(dep.PkgRefName, dep.ImportPath)
						}
					}
				}
				// 出参
				resp := m.Resp
				for _, dep := range resp.Deps {
					if dep.ImportPath != pkginfo.ImportPath { // 同包下无需import，否则循环依赖
						geni.AddImport(dep.PkgRefName, dep.ImportPath)
					}
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
					{{$LMethod.Name}}(req *{{ParamUnptrTypeInPkgInfo $LArg $LPkgInfo}}, stream {{$LSName}}_{{$LMethod.Name}}Server) (err error)
				{{- else}}
					{{$LMethod.Name}}(context.Context, *{{ParamUnptrTypeInPkgInfo $LArg $LPkgInfo}}) (*{{ParamUnptrTypeInPkgInfo $LResp $LPkgInfo}}, error)
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
					Recv() (*{{ParamUnptrTypeInPkgInfo $LArg $LPkgInfo}}, error)
					Send(*{{ParamUnptrTypeInPkgInfo $LResp $LPkgInfo}}) error
				}
			{{- else if $LMethod.ClientStreaming}}
				// {{$LSName}}_{{$LMethod.Name}}Server
				type {{$LSName}}_{{$LMethod.Name}}Server interface {
					streaming.Stream
		        	Recv() (*{{ParamUnptrTypeInPkgInfo $LArg $LPkgInfo}}, error)
		        	SendAndClose(*{{ParamUnptrTypeInPkgInfo $LResp $LPkgInfo}}) error	
				}
			{{- else if $LMethod.ServerStreaming}}
				// {{$LSName}}_{{$LMethod.Name}}Server
				type {{$LSName}}_{{$LMethod.Name}}Server interface {
					streaming.Stream
					Send(*{{ParamUnptrTypeInPkgInfo $LResp $LPkgInfo}}) error
				}
			{{- else}}
				{{/* 没事做 */}}
			{{- end}}
		{{- end}}


	{{- end}}
{{- end}}

`
