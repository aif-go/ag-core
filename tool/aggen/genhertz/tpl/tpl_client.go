package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
)

// ClientImportsSetter 设置Import部分信息
var ClientImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo
	// _pkgI := geni.PackageInfo
	// _svc := geni.ServiceInfo

	// FIXME 因为代码生成可能不是服务端应用，所以通过go_package获取import路径不准确，目前通过当前module名 + package名组合为import路径
	if _module.HasPwdGoMod {
		// 若当前存在go module则以当前gomodule路径为准
		geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	} else {
		geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	}

	geni.AddImports("context")
	geni.AddImport("config", "github.com/cloudwego/hertz/pkg/common/config")
	geni.AddImport("client", "github.com/aif-go/ag-core/ag/ag_hertz/client")

	return nil
}

// ClientTpl is the template for generating server.go.
var ClientTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	clientTpl +
	""

var clientTpl string = `
{{- $LGenI := .}}
{{- $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{- $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{- $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{- $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{- $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{- $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- $LPkgRefName := $LPkgInfo.PkgRefName }} {{/*包名*/}}
{{- $LServiceName := $LSI.ServiceName }}{{/*服务名*/}}
{{- $LServiceTypeName := call $LSI.ServiceTypeName}} {{/*服务类型名(接口)*/}}

{{- $LClientName := $LServiceName | printf "%sHertzClient"}}

// {{$LClientName}} is the hertz client for {{$LServiceName}}.
type {{$LClientName}} interface {
	{{- range $LSI.AllMethods}}
		{{- $LMethod := .}} {{/*方法*/}}
		{{- $LArgs := index $LMethod.Args 0}} {{/* 入参目前规定只有一个 */}}

		{{- if not $LMethod.IsStreaming }} {{/* 非流方法才可以构建http服务 */}}
			{{- if gt (len $LMethod.HttpDescs) 0}} {{/* 判断是否有http规则 */}}
				// {{$LMethod.Name}}
				{{$LMethod.Name}}(ctx context.Context, req {{$LArgs.Type}}, opts ...config.RequestOption) (resp *{{$LMethod.Resp.UnptrType}}, err error)
			{{- else}}
				// {{$LMethod.Name}} no http rule
			{{- end}}
		{{- else}}
			// {{$LMethod.Name}} is streaming method
		{{- end}}
	{{- end}}
}

// New{{$LClientName}} is the constructor for {{$LClientName}}.
func New{{$LClientName}}(hc *client.Client) {{$LClientName}} {
	return &{{$LClientName}}Impl{
		hc: hc,
	}
}

// {{$LClientName}}Impl is the hertz client impl for {{$LServiceName}}.
type {{$LClientName}}Impl struct {
	hc *client.Client
}


{{- range $LSI.AllMethods}}
	{{- $LMethod := .}}
	{{- $LArgs := index $LMethod.Args 0}} {{/* 入参目前规定只有一个 */}}
	{{- if not $LMethod.IsStreaming }} {{/* 非流方法才可以构建http服务 */}}
		{{- if gt (len $LMethod.HttpDescs) 0}} {{/* 判断是否有http规则 */}}
			{{- with index $LMethod.HttpDescs 0}} {{/* FIXME 目前只对主规则生成client调用*/}}
			    {{- $LHttpDesc := .}}
				// {{$LMethod.Name}} client impl {{$LMethod.Name}}
				func (c *{{$LClientName}}Impl) {{$LMethod.Name}}(ctx context.Context, req {{$LArgs.Type}}, opts ...config.RequestOption) (*{{$LMethod.Resp.UnptrType}}, error){
				    var resp {{$LMethod.Resp.UnptrType}}
				    path := "{{$LHttpDesc.Path}}"

					pathVars := make(map[string]string)
					{{- if $LHttpDesc.HasVars}}
					    {{- range $LHttpDesc.PathVars}}
						    pathVars["{{.}}"] = req.Get{{UpperFirst .}}()
					    {{- end}}
					{{- end}}

					{{- if $LHttpDesc.HasBody}}
						err := c.hc.Invoke(ctx, "{{$LHttpDesc.Method}}", path, pathVars, req{{$LHttpDesc.Body}}, &resp{{.ResponseBody}}, opts...)
					{{- else}}
						err := c.hc.Invoke(ctx, "{{$LHttpDesc.Method}}", path, pathVars, nil, &resp{{.ResponseBody}}, opts...)
					{{- end}}
					if err != nil {
						return nil, err
					}
					return &resp, nil
				}
			{{- end}}
		{{- end}}
	{{- else}}
		// {{$LMethod.Name}} is streaming method
	{{- end}}
{{- end}}

`
