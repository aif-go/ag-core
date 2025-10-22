package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// FxImportsSetter 设置Import部分信息
var FxImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo

	if !_module.HasPwdGoMod {
		return fmt.Errorf("pwdGoMod is empty")
	}

	// geni.AddImport("api", fmt.Sprintf("%s/%s", _module.PwdGoMod, "api"))

	geni.AddImport("fx", "go.uber.org/fx")
	geni.AddImport("hserver", "ag-core/contribute/aghertz/server")
	return nil
}

// FxTpl is the template for generating server.go.
var FxTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxTpl +
	""

var fxTpl string = `
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


var Fx{{$LServiceName}}HertzModule = fx.Module("fx_{{$LServiceName}}_hertz",
	{{- range $LSI.AllMethods}}
		{{- $LMethod := .}}
		{{- if not $LMethod.IsStreaming }} {{/* 非流方法才可以构建http服务 */}}
			{{- if gt (len $LMethod.HttpDescs) 0}} {{/* 判断是否有http规则 */}}
				Fx{{$LServiceName}}_{{$LMethod.Name}}_hertz_RegProvider,
			{{- end}}
		{{- end}}
	{{- end}}
)

{{- range $LSI.AllMethods}}
	{{- $LMethod := .}} {{/*方法*/}}
	{{- $LArgs := index $LMethod.Args 0}} {{/* 入参目前规定只有一个 */}}
	{{- if not $LMethod.IsStreaming }} {{/* 非流方法才可以构建http服务 */}}
		{{- if gt (len $LMethod.HttpDescs) 0}} {{/* 判断是否有http规则 */}}
			var Fx{{$LServiceName}}_{{$LMethod.Name}}_hertz_RegProvider =fx.Provide(
			{{- range $LMethod.HttpDescs}} 
			    {{- $LHttpDesc := .}}
					hserver.NewFxServerRouteProvider(
						Router_{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_{{.Method}}_Hertz,
					),
					{{/*
				    fx.Annotate(
						Register_{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_{{.Method}}_Hertz,
				        fx.ResultTags(` + "`" + `group:"hertz_router_options"` + "`" + `),
				    ),
					*/}}
			{{- end}}
			)
		{{- end}}
	{{- end}}
{{- end}}

`
