/*
agkitex fx注入部分模板
*/
package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
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
	geni.AddImport("akserver", "github.com/aif-go/ag-core/contribute/agkitex/server")

	return nil
}

// FxTpl is the template for generating server.go.

var FxTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxTpl +
	""

var fxTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}


{{- $LServiceName := $LSI.ServiceName }}{{/*服务名*/}}
{{- $LServiceTypeName := call $LSI.ServiceTypeName}} {{/*服务类型名(接口)*/}}

// Fx{{$LServiceName}}KitexModule fx module
var Fx{{$LServiceName}}KitexModule = fx.Module("fx_{{$LServiceName}}_kitex",
	Fx{{$LServiceName}}KitexRegProvider,
)

// Fx{{$LServiceName}}KitexRegProvider fx provide
var Fx{{$LServiceName}}KitexRegProvider = fx.Provide(

	{{/* 
	fx.Annotate(
		Register_{{$LServiceName}}_KitexServer,
		fx.ResultTags(` + "`" + `group:"ag_kitex_server_registrars"` + "`" + `),
	),
	*/}}

	akserver.NewFxAgKitexServiceRegistry(
		Register_{{$LServiceName}}_KitexServer,
	),
)

`
