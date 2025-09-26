package genserver

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/genserver/tpl"
	"ag-core/tool/aggen/types"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// GenServiceTask 生成服务任务
func GenServiceTask() *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	// 添加PackageGroup级别的任务生成器
	genners.AddGen(types.ScopePackageGroup, PackageGroupTaskGen)

	// 添加module级别的任务生成器
	genners.AddGen(types.ScopeModule, ModuleTaskGen)

	return genners
}

var PackageGroupTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	_pkginfo := geni.PkgInfo
	_pkgGroup := geni.PackageGroup

	if _pkgGroup == nil {
		return nil, fmt.Errorf("packageGroup is nil")
	}

	pkgname := _pkginfo.PkgName
	apipath := strings.ReplaceAll(pkgname, ".", string(filepath.Separator)) // 将pkg名替换为系统文件路径

	infName := fmt.Sprintf("%s_%s_%s", "agserver", _pkginfo.PkgRefName, "interface.go")
	infTask := &types.Task{
		Name:      infName,
		Path:      path.Join(apipath, infName),
		Text:      tpl.InterfaceTpl,
		SetImport: tpl.InterfaceImportsSetter,
	}
	tasks = append(tasks, infTask)

	return tasks, nil
}

var ModuleTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	// _module := geni.ModuleInfo

	// adpPath := "api"
	adpPath := "internal/adpgen"
	fxAdapterName := "zfx_adapter.go"
	fxAdapterTask := &types.Task{
		Name:      fxAdapterName,
		Path:      path.Join(adpPath, fxAdapterName),
		Text:      tpl.FxAdapterTpl,
		SetImport: tpl.FxAdapterImportsSetter,
	}
	tasks = append(tasks, fxAdapterTask)

	adpInitPath := path.Join(adpPath, "adpinit")
	fxAdapterInitName := "zfx_adapter_init.go"
	fxAdapterInitTask := &types.Task{
		Name:      fxAdapterInitName,
		Path:      path.Join(adpInitPath, fxAdapterInitName),
		Text:      tpl.FxAdapterInitTpl,
		SetImport: tpl.FxAdapterInitImportsSetter,
	}
	tasks = append(tasks, fxAdapterInitTask)

	return tasks, nil
}
