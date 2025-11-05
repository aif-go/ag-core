package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/genkitex/tpl/kitextpl"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ServerImportsSetter 设置Import部分信息
var ServerImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo
	// _pkgI := geni.PackageInfo
	// _svc := geni.ServiceInfo

	if _module.HasPwdGoMod {
		// 若当前存在go module则以当前gomodule路径为准
		geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	} else {
		geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	}

	// geni.AddImports("server") // 注意全局依赖的配置
	geni.AddImport("server", "github.com/cloudwego/kitex/server")

	// geni.AddImport("akserver", "ag-core/ag/ag_kitex/server")
	geni.AddImport("akserver", "ag-core/contribute/agkitex/server")

	return nil
}

// ServerTpl is the template for generating server.go.
var ServerTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	serverTpl +
	""

var serverTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}


{{- with $LSI }}
` +
	kitextpl.ServerTpl + // 基础kitex server代码
	agkitexServerTpl_ext1 +
	`
{{ end }}
`

// agkitexServerTpl_ext1 扩展1，添加ag_kitex的服务注册方式
var agkitexServerTpl_ext1 string = `
// Register_{{.ServiceName}}_KitexServer ag_kitex service register.
{{/* 
func Register_{{.ServiceName}}_KitexServer(srv {{call .ServiceTypeName}}) akserver.Option {
	return akserver.WithServiceRegistrar(&akserver.ServiceRegistrar{
		ServiceInfo: NewServiceInfo(),
		Handler: srv,
	})
}
*/}}
func Register_{{.ServiceName}}_KitexServer(srv {{call .ServiceTypeName}}) *akserver.AgKitexServiceRegistry {
	reg := akserver.NewAgKitexServiceRegistry(
		NewServiceInfo(),
		srv,
	)
	return reg
}
`
