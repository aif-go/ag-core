package tpl

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
)

// module 级别
// FxAdapterImportsSetter 设置Import部分信息
var FxAdapterImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	if !_module.HasPwdGoMod {
		return fmt.Errorf("pwdGoMod is empty")
	}
	// geni.AddImport("adpinit", fmt.Sprintf("%s/%s/%s", _module.PwdGoMod, "api", "adpinit"))
	geni.AddImport("adpinit", fmt.Sprintf("%s/%s/%s/%s", _module.PwdGoMod, "internal", "adpgen", "adpinit"))

	geni.AddImport("fx", "go.uber.org/fx")
	return nil
}

// FxAdapterTpl is the template for generating server.go.
var FxAdapterTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	fxAdapterTpl +
	""

var fxAdapterTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}


// FxAdapterModule return adapter fx module
func FxAdapterModule() fx.Option {
	fxAdapterOpts := adpinit.GetFxAdapterOpt()
	return fx.Module("fx-adapter", fxAdapterOpts...)
}

`
