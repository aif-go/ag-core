package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
)

// PackageGroup级别
var FxProxyIsSkip_v2 = func(geni *types.GennerInfo) bool {
	pkgg := geni.PackageGroup

	for _, pkg := range pkgg.PackageInfos {
		if len(pkg.Services) > 0 {
			return false
		}
	}
	return true
}

// FxProxyImportsSetter 设置Import部分信息
var FxProxyImportsSetter_v2 = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo

	geni.AddImport("fx", "go.uber.org/fx")

	// geni.AddImport("smw", "github.com/aif-go/ag-core/ag/ag_ext")
	// geni.AddImport("smw", "github.com/aif-go/ag-core/ag/ag_service")
	geni.AddImport("ag_service", "github.com/aif-go/ag-core/ag/ag_service")

	serviceRef := fmt.Sprintf("%s/internal/service", _module.PwdGoMod)
	geni.AddImport("service", serviceRef)

	// 包组级别，proxy依赖接口api包名
	geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath) // 此仅导入api包路径

	return nil
}

// FxProxyTpl is the template for generating the servicename.go source.
var FxProxyTpl_v2 string = "" +
	generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxProxyTpl_v2 +
	""

var fxProxyTpl_v2 string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{/* {{ $LPI := .PackageInfo */}}    {{/* IDL文件级别信息 当前在packageGroup级别下，不存在 */}}
{{/* $LSI := .ServiceInfo */}}    {{/* 服务信息 当前在packageGroup级别下，不存在 */}}


{{- $LPkgRefName := $LPkgInfo.PkgRefName}}

{{- range $LPI := $LPG.PackageInfos }} {{/* 包组信息 */}}
	{{- range $LSIindex, $LSI := $LPI.Services}}
		{{- $LServiceName := $LSI.ServiceName}}
		{{- $LLowerServiceName := ToLower $LServiceName}}
		{{- $LServiceImplName := printf "%sImpl" $LSI.ServiceName}}
		
		// FxIn{{$LServiceName}}Middleware servcie middleware inject params
		type FxIn{{$LServiceName}}Middleware struct {
			fx.In

			AgServiceBuilder ag_service.AgServiceBuilder

			CustomMws []ag_service.MiddlewareProvider ` + "`" + `group:"fx_{{$LLowerServiceName}}_service_middleware" ,optional:"true"` + "`" + `
		}
		
		// New{{$LServiceName}}ProxyWithFxIn create {{$LServiceName}} proxy with fx inject
		func New{{$LServiceName}}ProxyWithFxIn(in FxIn{{$LServiceName}}Middleware, svc *service.{{$LServiceImplName}}) ({{$LPkgRefName}}.{{$LServiceName}}, error) {
			return New{{$LServiceName}}Proxy(svc, in.AgServiceBuilder, in.CustomMws)
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
