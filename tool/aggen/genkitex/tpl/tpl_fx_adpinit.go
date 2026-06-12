/*
agkitex fx注入部分模板
*/
package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
	"strings"
)

// FxAdpInitImportsSetter 设置Import部分信息
var FxAdpInitImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	if !_module.HasPwdGoMod {
		return fmt.Errorf("pwdGoMod is empty")
	}

	// _pkgInfo := geni.PkgInfo
	_svcInfo := geni.ServiceInfo
	lowerSvcName := strings.ToLower(_svcInfo.ServiceName)
	//eg: import hzwhello "ag-layout-demo/api/hzw/hzwhello
	// geni.AddImport(lowerSvcName, fmt.Sprintf("%s/%s/%s", _module.PwdGoMod, _pkgInfo.ImportPkg, lowerSvcName))

	//eg: import hzwhello "ag-layout-demo/internal/adpgen/kitex/hzwhello"
	geni.AddImport(lowerSvcName, fmt.Sprintf("%s/%s/%s/%s/%s", _module.PwdGoMod, "internal", "adpgen", "kitex", lowerSvcName))

	return nil
}

// FxAdpInitTpl is the template for generating server.go.

var FxAdpInitTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxAdpInitTpl +
	""

var fxAdpInitTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- $LServiceName := $LSI.ServiceName }}{{/*服务名*/}}
func init(){
	AddFxAdapterOpt({{ToLower $LServiceName}}.Fx{{$LServiceName}}KitexModule)
}
`
