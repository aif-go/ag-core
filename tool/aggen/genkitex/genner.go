package genkitex

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/genkitex/tpl"
	"ag-core/tool/aggen/types"
	"fmt"
	"path"
	"strings"
)

// KitexGenServiceTask 生成服务任务
func KitexGenServiceTask() *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	genners.AddGen(types.ScopeService, ServiceTaskGen)

	return genners
}

var ServiceTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	err := geni.CheckServiceScop()
	if err != nil {
		return nil, err
	}

	_pkgInfo := geni.PkgInfo
	_svcInfo := geni.ServiceInfo

	// apipath := strings.ReplaceAll(_pkgInfo.PkgName, ".", string(filepath.Separator))
	lowerSvcName := strings.ToLower(_svcInfo.ServiceName)
	// outpath := path.Join(apipath, lowerSvcName) // xxx/api/hzw/hello
	outpath := path.Join("internal", "adpgen", "kitex", lowerSvcName)

	// servicefile
	svcName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, ".go")
	svcTask := &types.Task{
		Name:      svcName,
		Path:      path.Join(outpath, svcName),
		Text:      tpl.ServiceTpl,
		SetImport: tpl.ServiceImportsSetter,
	}
	tasks = append(tasks, svcTask)

	// serverfile FIXME 属于适配层的
	svrName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, "_server.go")
	svrTask := &types.Task{
		Name:      svrName,
		Path:      path.Join(outpath, svrName),
		Text:      tpl.ServerTpl,
		SetImport: tpl.ServerImportsSetter,
	}
	tasks = append(tasks, svrTask)

	// clientfile
	cliName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, "_client.go")
	cliTask := &types.Task{
		Name:      cliName,
		Path:      path.Join(outpath, cliName),
		Text:      tpl.ClientTpl,
		SetImport: tpl.ClientImportsSetter,
	}
	tasks = append(tasks, cliTask)

	// fxfile
	fxName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, "_fx.go")
	fxTask := &types.Task{
		Name:      fxName,
		Path:      path.Join(outpath, fxName),
		Text:      tpl.FxTpl,
		SetImport: tpl.FxImportsSetter,
	}
	tasks = append(tasks, fxTask)

	// fxadpinitfile
	fxAdpInitName := fmt.Sprintf("%s_%s_%s%s", "zfx_agkitex", _pkgInfo.PkgRefName, lowerSvcName, "_adpinit.go")
	fxAdpInitPath := path.Join("internal", "adpgen", "adpinit")
	fxAdpInitTask := &types.Task{
		Name:      fxAdpInitName,
		Path:      path.Join(fxAdpInitPath, fxAdpInitName),
		Text:      tpl.FxAdpInitTpl,
		SetImport: tpl.FxAdpInitImportsSetter,
	}
	tasks = append(tasks, fxAdpInitTask)

	return tasks, nil
}
