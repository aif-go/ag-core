package tpl

import (
	"ag-core/tool/aggen/genkitex/tpl/kitextpl"
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ClientImportsSetter 设置Import部分信息
var ClientImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo
	// _pkgI := geni.PackageInfo
	_svc := geni.ServiceInfo

	if _module.HasPwdGoMod {
		// 若当前存在go module则以当前gomodule路径为准
		geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	} else {
		geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	}

	geni.AddImport("client", "github.com/cloudwego/kitex/client")

	if _svc.HasStreaming {
		geni.AddImport("streaming", "github.com/cloudwego/kitex/pkg/streaming")
		geni.AddImport("transport", "github.com/cloudwego/kitex/transport")
		geni.AddImport("streamcall", "github.com/cloudwego/kitex/client/callopt/streamcall")
		geni.AddImport("streamclient", "github.com/cloudwego/kitex/client/streamclient")
	}
	if len(_svc.AllMethods()) > 0 {
		// TODO 暂忽略NeedCallOpt判断
		// if aggen.NeedCallOpt(pkg) {
		// 	// pkg.AddImports("callopt")
		// 	pkg.AddImport("callopt", "github.com/cloudwego/kitex/client/callopt")
		// }
		// // pkg.AddImports("context")
		// pkg.AddImport("context", "context")
		geni.AddImport("callopt", "github.com/cloudwego/kitex/client/callopt")

		geni.AddImport("context", "context")
	}

	return nil
}

// ClientTpl is the template for generating client.go.
var ClientTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	clientTpl +
	""

var clientTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- with $LSI }}
` +
	kitextpl.ClientTpl + // 基础kitex client代码
	// clienttest + // 基础kitex client代码
	`
{{ end }}
`
var clienttest string = ``
