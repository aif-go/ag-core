package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
)

// PackageGroup级别
var FxServiceIsSkip = func(geni *types.GennerInfo) bool {
	pkgg := geni.PackageGroup

	for _, pkg := range pkgg.PackageInfos {
		if len(pkg.Services) > 0 {
			return false
		}
	}
	return true
}

// FxServiceImportsSetter 设置Import部分信息
var FxServiceImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	// _pkgI := geni.PackageInfo
	// _svc := geni.ServiceInfo

	geni.AddImport("fx", "go.uber.org/fx")

	if !_module.HasPwdGoMod {
		return fmt.Errorf("pwdGoMod is empty")
	}

	serviceRef := fmt.Sprintf("%s/internal/service", _module.PwdGoMod)
	geni.AddImport("service", serviceRef) // 模块servcie import

	return nil
}

// FxServiceTpl is the template for generating the servicename.go source.
var FxServiceTpl string = "" +
	generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxServiceTpl +
	""

var fxServiceTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{/* {{ $LPI := .PackageInfo */}}    {{/* IDL文件级别信息 当前在packageGroup级别下，不存在 */}}
{{/* $LSI := .ServiceInfo */}}    {{/* 服务信息 当前在packageGroup级别下，不存在 */}}

{{- $LPkgRefName := $LPkgInfo.PkgRefName}}

{{- /*
var Fx{{$LPkgRefName}}ServiceModule = fx.Module("fx-{{$LPkgRefName}}-service",
	fx.Provide(
		{{- range $LPI := $LPG.PackageInfos }} 
			{{- range $LSI := $LPI.Services}}
				{{- $LServiceImplName := printf "%sImpl" $LSI.ServiceName}}
				service.New{{$LServiceImplName}}, // constructor for {{$LServiceImplName}}
			{{- end}}
		{{- end}}
	),
)
*/}}

func init() {
	AddFxServiceOpt(
		fx.Provide(
			{{- range $LPI := $LPG.PackageInfos }} {{/* 包组信息 */}}
				{{- range $LSI := $LPI.Services}}
					{{- $LServiceImplName := printf "%sImpl" $LSI.ServiceName}}
					service.New{{$LServiceImplName}}I, // constructor return interface
				{{- end}}
			{{- end}}
		),
	)
}
`
