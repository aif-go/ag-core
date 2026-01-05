package genapi

import (
	"ag-core/tool/aggen/genapi/tpl"
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
	"path"
)

// GenApiTask 生成接口任务
func GenApiTask() *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	// 添加PackageGroup级别的任务生成器
	genners.AddGen(types.ScopePackageGroup, PackageGroupBaseTaskGen)

	return genners
}

var PackageGroupBaseTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	_pkginfo := geni.PkgInfo
	_pkgGroup := geni.PackageGroup

	if _pkgGroup == nil {
		return nil, fmt.Errorf("packageGroup is nil")
	}

	// pkgname := _pkginfo.PkgName
	// apipath := strings.ReplaceAll(pkgname, ".", string(filepath.Separator)) // 将pkg名替换为系统文件路径
	apipath := fmt.Sprintf("%s/%s", "api", _pkginfo.PkgRefName) // api的目录

	infName := fmt.Sprintf("%s_%s_%s", "agserver", _pkginfo.PkgRefName, "interface.go")
	infTask := &types.Task{
		Name:      infName,
		Path:      path.Join(apipath, infName),
		Text:      tpl.InterfaceTpl,
		SetImport: tpl.InterfaceImportsSetter,
		IsSkip:    tpl.InterfaceIsSkip,
	}
	tasks = append(tasks, infTask)

	return tasks, nil
}
