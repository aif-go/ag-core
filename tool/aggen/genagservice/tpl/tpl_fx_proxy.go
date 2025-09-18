package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// PackageGroup级别

// FxProxyImportsSetter 设置Import部分信息
var FxProxyImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo

	geni.AddImport("fx", "go.uber.org/fx")

	geni.AddImport("smw", "ag-core/ag/ag_ext")

	serviceRef := fmt.Sprintf("%s/internal/service", _module.PwdGoMod)
	geni.AddImport("service", serviceRef)

	if _module.HasPwdGoMod {
		// 若当前存在go module则以当前gomodule路径为准
		geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	} else {
		geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	}
	return nil
}

// FxProxyTpl is the template for generating the servicename.go source.
var FxProxyTpl string = "" +
	generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxProxyTpl +
	""

var fxProxyTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{/* {{ $LPI := .PackageInfo */}}    {{/* IDL文件级别信息 当前在packageGroup级别下，不存在 */}}
{{/* $LSI := .ServiceInfo */}}    {{/* 服务信息 当前在packageGroup级别下，不存在 */}}


{{- $LPkgRefName := $LPkgInfo.PkgRefName}}


var Fx{{$LPkgRefName}}ServiceProxyModule = fx.Module("fx-{{$LPkgRefName}}-service-proxy",
	fx.Provide(
		{{- range $LPI := $LPG.PackageInfos }} {{/* 包组信息 */}}
			{{- range $LSIindex, $LSI := $LPI.Services}}
				{{- $LServiceName := $LSI.ServiceName}}
				{{- $LServiceImplName := printf "%sImpl" $LSI.ServiceName}}
				New{{$LServiceName}}ProxyWithFxIn,	
			{{- end}}
		{{- end}}
	),
)

{{- range $LPI := $LPG.PackageInfos }} {{/* 包组信息 */}}
	{{- range $LSIindex, $LSI := $LPI.Services}}
		{{- $LServiceName := $LSI.ServiceName}}
		{{- $LLowerServiceName := ToLower $LServiceName}}
		{{- $LServiceImplName := printf "%sImpl" $LSI.ServiceName}}
		
		// FxIn{{$LServiceName}}Middleware servcie middleware inject params
		type FxIn{{$LServiceName}}Middleware struct {
			smw.BaseFxMiddlewareParams
			CustomMws []smw.PrioritizedMiddleware ` + "`" + `group:"fx_{{$LLowerServiceName}}_service_middleware" ,optional:"true"` + "`" + `
		}
		
		// New{{$LServiceName}}ProxyWithFxIn create {{$LServiceName}} proxy with fx inject
		func New{{$LServiceName}}ProxyWithFxIn(in FxIn{{$LServiceName}}Middleware, svc *service.{{$LServiceImplName}}) {{$LPkgRefName}}.{{$LServiceName}} {
			return New{{$LServiceName}}Proxy(svc,append(in.GlobalMws,in.CustomMws...))
		}
	{{- end}}
{{- end}}

func init() {
	AddFxServiceWithProxyOpt(
		fx.Provide(

			{{- range $LPI := $LPG.PackageInfos }} {{/* 包组信息 */}}
				{{- range $LSIindex, $LSI := $LPI.Services}}
					{{- $LServiceName := $LSI.ServiceName}}
					{{- $LLowerServiceName := ToLower $LServiceName}}
					{{- $LServiceImplName := printf "%sImpl" $LSI.ServiceName}}
					service.New{{$LServiceImplName}}, // constructor return impl
					New{{$LServiceName}}ProxyWithFxIn,
				{{- end}}
			{{- end}}
		),
	)
}


`
