package genagservice

import (
	"ag-core/tool/aggen/genagservice/tpl"
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
	"path"
	"strings"
)

// GenServiceTask 生成服务任务
func GenServiceTask() *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	// service级别
	genners.AddGen(types.ScopeService, ServiceScopeTaskGen)

	// packageGroup级别
	genners.AddGen(types.ScopePackageGroup, PackageGroupScopeTaskGen)

	// module级别
	genners.AddGen(types.ScopeModule, ModuleScopeTaskGen)

	return genners
}

// ServiceScopeTaskGen 服务级别任务生成
var ServiceScopeTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	// _moduleInfo := geni.ModuleInfo
	// _pkginfo := geni.PkgInfo
	// _pkgGroup := geni.PackageGroup
	_svcInfo := geni.ServiceInfo

	lowerSvcName := strings.ToLower(_svcInfo.ServiceName)

	// service
	// svcPath := path.Join(_moduleInfo.PwdGoModPath, "internal", "service") // xxx/internal/service
	svcPath := path.Join("internal", "service") // xxx/internal/service
	svcName := fmt.Sprintf("%s_%s%s", "agservice", lowerSvcName, ".go")
	svcTask := &types.Task{
		Name:       svcName,
		Path:       path.Join(svcPath, svcName),
		Text:       tpl.ServiceTpl,
		SetImport:  tpl.ServiceImportsSetter,
		OutputType: types.OutputTypeNotOverride, // 生成代码 不覆盖
		// OutputType: types.OutputTypeOverride, // TODO 测试时方便，覆盖
	}
	tasks = append(tasks, svcTask)

	// proxy
	proxyPath := path.Join("internal", "svcgen")
	proxyName := fmt.Sprintf("%s_%s%s", "agservice", lowerSvcName, "_proxy.go")
	proxyTask := &types.Task{
		Name:      proxyName,
		Path:      path.Join(proxyPath, proxyName),
		Text:      tpl.ProxyTpl_v2,
		SetImport: tpl.ProxyImportsSetter_v2,
	}
	tasks = append(tasks, proxyTask)

	return tasks, nil
}

// PackageGroupScopeTaskGen 包组级别任务生成
var PackageGroupScopeTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	// _moduleInfo := geni.ModuleInfo
	_pkginfo := geni.PkgInfo
	// _pkgGroup := geni.PackageGroup

	lowerPkgRefName := strings.ToLower(_pkginfo.PkgRefName)

	// fx_service // FIXME 暂不生成非代理服务注册
	// fxSvcPath := path.Join("internal", "svcgen")
	// fxSvcName := fmt.Sprintf("%s_%s%s", "zfx_agservice_service", lowerPkgRefName, ".go")
	// fxSvcTask := &types.Task{
	// 	Name:      fxSvcName,
	// 	Path:      path.Join(fxSvcPath, fxSvcName),
	// 	Text:      tpl.FxServiceTpl,
	// 	SetImport: tpl.FxServiceImportsSetter,
	// 	IsSkip:    tpl.FxServiceIsSkip,
	// }
	// tasks = append(tasks, fxSvcTask)

	// fx_proxy
	fxProxyPath := path.Join("internal", "svcgen")
	fxProxyName := fmt.Sprintf("%s_%s%s", "zfx_agservice_proxy", lowerPkgRefName, ".go")
	fxProxyTask := &types.Task{
		Name: fxProxyName,
		Path: path.Join(fxProxyPath, fxProxyName),
		// Text:      tpl.FxProxyTpl,
		// SetImport: tpl.FxProxyImportsSetter,
		// IsSkip:    tpl.FxProxyIsSkip,
		Text:      tpl.FxProxyTpl_v2,
		SetImport: tpl.FxProxyImportsSetter_v2,
		IsSkip:    tpl.FxProxyIsSkip_v2,
	}
	tasks = append(tasks, fxProxyTask)

	return tasks, nil
}

// ModuleScopeTaskGen 模块级别任务生成
var ModuleScopeTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	// zfx_global_service
	fxGlobalSvcPath := path.Join("internal", "svcgen")
	fxGlobalSvcName := fmt.Sprintf("%s%s", "zfx", "_service.go")
	fxGlobalSvcTask := &types.Task{
		Name:      fxGlobalSvcName,
		Path:      path.Join(fxGlobalSvcPath, fxGlobalSvcName),
		Text:      tpl.FxGlobalServiceTpl,
		SetImport: tpl.FxGlobalServiceImportsSetter,
	}
	tasks = append(tasks, fxGlobalSvcTask)

	return tasks, nil
}
