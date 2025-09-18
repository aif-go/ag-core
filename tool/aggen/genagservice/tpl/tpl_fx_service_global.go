package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
)

// PackageGroup级别

// FxServiceImportsSetter 设置Import部分信息
var FxGlobalServiceImportsSetter = func(geni *types.GennerInfo) error {

	geni.AddImport("fx", "go.uber.org/fx")
	geni.AddImports("sync")

	return nil
}

// FxServiceTpl is the template for generating the servicename.go source.
var FxGlobalServiceTpl string = "" +
	generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxGlobalServiceTpl +
	""

var fxGlobalServiceTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}

var (
	_lock1             sync.Mutex
	fxServiceOpts      []fx.Option
	_lock2             sync.Mutex
	fxServiceWithProxyOpts []fx.Option
)


// FxServiceModule return service fx module
func FxServiceModule() fx.Option {
	_lock1.Lock()
	defer _lock1.Unlock()
	return fx.Module("fx-service", fxServiceOpts...)
}

// FxServiceWithProxyModule return service fx module with service proxy
func FxServiceWithProxyModule() fx.Option {
	_lock2.Lock()
	defer _lock2.Unlock()
	return fx.Module("fx-service-with-proxy", fxServiceWithProxyOpts...)
}

// AddFxServiceOpt add fx service option.
func AddFxServiceOpt(opt fx.Option) {
	_lock1.Lock()
	defer _lock1.Unlock()
	if fxServiceOpts == nil {
		fxServiceOpts = []fx.Option{}
	}
	fxServiceOpts = append(fxServiceOpts, opt)
}

// AddFxServiceWithProxyOpt add fx service option with proxy.
func AddFxServiceWithProxyOpt(opt fx.Option) {
	_lock2.Lock()
	defer _lock2.Unlock()
	if fxServiceWithProxyOpts == nil {
		fxServiceWithProxyOpts = []fx.Option{}
	}
	fxServiceWithProxyOpts = append(fxServiceWithProxyOpts, opt)
}
`
