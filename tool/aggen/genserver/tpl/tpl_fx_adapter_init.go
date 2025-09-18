package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
)

// module 级别
// FxAdapterInitImportsSetter 设置Import部分信息
var FxAdapterInitImportsSetter = func(geni *types.GennerInfo) error {
	geni.AddImport("fx", "go.uber.org/fx")
	geni.AddImports("sync")

	return nil
}

// FxAdapterTpl is the template for generating server.go.
var FxAdapterInitTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxAdapterInitTpl +
	""

var fxAdapterInitTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}

var (
	fxAdapterLock        sync.Mutex
	fxAdapterOpts []fx.Option
)

// AddFxAdapterOpt add fx adapter option.
func AddFxAdapterOpt(opt fx.Option) {
	fxAdapterLock.Lock()
	defer fxAdapterLock.Unlock()
	if fxAdapterOpts == nil {
		fxAdapterOpts = []fx.Option{}
	}
	fxAdapterOpts = append(fxAdapterOpts, opt)
}

// GetFxAdapterOpt .
func GetFxAdapterOpt() []fx.Option {
	fxAdapterLock.Lock()
	defer fxAdapterLock.Unlock()
	return fxAdapterOpts
}
`
