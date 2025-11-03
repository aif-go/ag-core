package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ClientImportsSetterV2 设置Import部分信息
var ClientImportsSetterV2 = func(geni *types.GennerInfo) error {
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
	geni.AddImports("github.com/cloudwego/hertz/pkg/app/client")
	geni.AddImport("config", "github.com/cloudwego/hertz/pkg/common/config")
	geni.AddImport("agclient", "ag-core/contribute/aghertz/aghertzclient")

	return nil
}

// ClientTpl is the template for generating server.go.
var ClientTplV2 string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	clientTplV2 +
	""

var clientTplV2 string = `
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
func New{{$LClientName}}(hc *client.Client, opts ...agclient.AgHertzClientOption) {{$LClientName}} {
	client := &{{$LClientName}}Impl{
		HertzBaseClient: agclient.NewHertzBaseClient(hc, opts...),
	}
	return client
}

// {{$LClientName}}Impl is the hertz client impl for {{$LServiceName}}.
type {{$LClientName}}Impl struct {
	*agclient.HertzBaseClient
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

					// 构建请求配置
					reqParam := &agclient.RequestParam{
						Method: "{{$LHttpDesc.Method}}",
						Path:   "{{$LHttpDesc.Path}}",
						{{- if $LHttpDesc.HasVars}}
						PathVars: map[string]string{
						    {{- range $LHttpDesc.PathVars}}
							    "{{.}}": req.Get{{UpperFirst .}}(),
						    {{- end}}
						},
						{{- end}}
					}

					{{- if $LHttpDesc.HasBody}}
						rerr := c.DoRequest(ctx, reqParam, req{{$LHttpDesc.Body}}, &resp{{.ResponseBody}}, opts...)
					{{- else}}
						rerr := c.DoRequest(ctx, reqParam, nil, &resp{{.ResponseBody}}, opts...)
					{{- end}}
						if rerr != nil {
							return nil, rerr
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
