package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
)

// ClientImportsSetter 设置Import部分信息
var AgClientImportsSetter = func(geni *types.GennerInfo) error {

	geni.AddImport("agclient", "github.com/aif-go/ag-core/contribute/agkitex/client")
	geni.AddImport("client", "github.com/cloudwego/kitex/client")

	return nil
}

// ClientTpl is the template for generating client.go.
var AgClientTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	agClientTpl +
	""

var agClientTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- with $LSI }}
	// NewClientWithSuite Create a new client with the given suite.
	func NewClientWithSuite(destService string, suite *agclient.KitexClientSuite, opts ...client.Option) (Client, error) {
		allOpts := append([]client.Option{client.WithSuite(suite)}, opts...)
		return NewClient(destService, allOpts...)
	}
{{ end }}
`
