package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ServiceImportsSetter 设置Import部分信息
var ServiceImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo
	// _pkgI := geni.PackageInfo
	// _svc := geni.ServiceInfo

	geni.AddImports("context")

	if _module.HasPwdGoMod {
		// 若当前存在go module则以当前gomodule路径为准
		geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	} else {
		geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	}

	return nil
}

// ServiceTpl is the template for generating the servicename.go source.
var ServiceTpl string = shoudEditTpl +
	generator.Tpl_pkg +
	generator.Tpl_import +
	serviceTpl +
	""

var shoudEditTpl string = `
	// TODO 这里需要根据业务场景，自行修改实现
`

var serviceTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}


{{- with $LSI }}
` + serviceInnerTpl + `
{{ end }}
`

var serviceInnerTpl string = `
{{- $LPkgRefName := $LPkgInfo.PkgRefName}}
{{- $LServiceInfo := .}}
{{- $LServiceName := $LServiceInfo.ServiceName}}
{{- $LServiceImplName := printf "%sImpl" $LServiceInfo.ServiceName}}


// {{$LServiceImplName}} defines the service implementation for {{$LServiceImplName}}.
type {{$LServiceImplName}} struct {
	// TODO 这里需要根据业务场景，添加依赖
}

// New{{$LServiceImplName}} creates and returns a new {{$LServiceImplName}} instance.
//@param TODO inject dependencies
//@return *{{$LServiceImplName}}
func New{{$LServiceImplName}}( /* TODO inject dependencies */ ) *{{$LServiceImplName}} {
	return &{{$LServiceImplName}}{}
}

// New{{$LServiceImplName}}I new {{$LServiceImplName}} instance, and return interface {{$LPkgRefName}}.{{$LServiceName}}.
//@param TODO inject dependencies
//@return {{$LPkgRefName}}.{{$LServiceName}}
func New{{$LServiceImplName}}I( /* TODO inject dependencies */ ) {{$LPkgRefName}}.{{$LServiceName}} {
	return New{{$LServiceImplName}}( /* TODO inject dependencies */ )
}

{{- range $LServiceInfo.AllMethods}}
	{{- $LMethod := .}}
{{- end}}


{{- range $LServiceInfo.AllMethods}}
	{{- $LMethod := .}}
	{{- $LArg := index $LMethod.Args 0 }} {{/* 规定接口参数只有一个参数 */}}
	{{- $LResp := $LMethod.Resp}}

	{{- if and $LMethod.ClientStreaming $LMethod.ServerStreaming }}
		// {{$LMethod.Name}} TODO:DESCRIBE
		// ClientStreaming and ServerStreaming
		func (c *{{$LServiceImplName}}) {{$LMethod.Name}}(stream {{$LPkgRefName}}.{{$LServiceName}}_{{$LMethod.Name}}Server) error{
			// TODO
			return nil
		}
	{{- else if $LMethod.ClientStreaming}}
		// {{$LMethod.Name}} TODO:DESCRIBE
		// ClientStreaming
		func (c *{{$LServiceImplName}}) {{$LMethod.Name}}(stream {{$LPkgRefName}}.{{$LServiceName}}_{{$LMethod.Name}}Server) error{
			// TODO
			return nil		
		}
	{{- else if $LMethod.ServerStreaming}}
		// {{$LMethod.Name}} TODO:DESCRIBE
		// ServerStreaming
		func (c *{{$LServiceImplName}}) {{$LMethod.Name}}(req {{$LArg.Type}}, stream {{$LPkgRefName}}.{{$LServiceName}}_{{$LMethod.Name}}Server) (error){
			// TODO
		    return nil
		}

	{{- else}}
		// {{$LMethod.Name}} TODO:DESCRIBE
		func (c *{{$LServiceImplName}}) {{$LMethod.Name}}(ctx context.Context, req {{$LArg.Type}}) ({{$LResp.Type}}, error){
			var resp *{{$LMethod.Resp.UnptrType}}
			resp = &{{$LMethod.Resp.UnptrType}}{
			}
			// TODO 添加业务处理逻辑

			return resp, nil
		}
	{{- end}}

{{- end}}

`
