package genserver

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/genserver/tpl"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"path"
)

// GenServiceTask 生成服务任务
func GenServiceTask(model string) *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	// 添加PackageGroup级别的任务生成器 FIXME api的生成移至genapi中
	// genners.AddGen(types.ScopePackageGroup, PackageGroupBaseTaskGen)

	// if model == "all" || model == "server" {
	// 添加module级别Server端的任务生成器
	genners.AddGen(types.ScopeModule, ModuleServerTaskGen)
	// }

	return genners
}

// FIXME api的生成移至genapi中
// var PackageGroupBaseTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
// 	tasks := make([]*types.Task, 0)

// 	_pkginfo := geni.PkgInfo
// 	_pkgGroup := geni.PackageGroup

// 	if _pkgGroup == nil {
// 		return nil, fmt.Errorf("packageGroup is nil")
// 	}

// 	// pkgname := _pkginfo.PkgName
// 	// apipath := strings.ReplaceAll(pkgname, ".", string(filepath.Separator)) // 将pkg名替换为系统文件路径
// 	apipath := fmt.Sprintf("%s/%s", "api", _pkginfo.PkgRefName) // api的目录

// 	infName := fmt.Sprintf("%s_%s_%s", "agserver", _pkginfo.PkgRefName, "interface.go")
// 	infTask := &types.Task{
// 		Name:      infName,
// 		Path:      path.Join(apipath, infName),
// 		Text:      tpl.InterfaceTpl,
// 		SetImport: tpl.InterfaceImportsSetter,
// 		IsSkip:    tpl.InterfaceIsSkip,
// 	}
// 	tasks = append(tasks, infTask)

// 	return tasks, nil
// }

var ModuleServerTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
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
