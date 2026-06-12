package genkitex

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/genkitex/tpl"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
	"path"
	"strings"
)

// KitexGenServiceTask 生成服务任务
func KitexGenServiceTask(model string) *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	genners.AddGen(types.ScopeService, ServiceBaseTaskGen)

	if model == "all" || model == "server" {
		genners.AddGen(types.ScopeService, ServiceServerTaskGen)
	}

	// FIXME kitex原生成逻辑未区分clien和server端，server端生成文件中存在引用client内容的耦合情况，此处暂都生成client端文件
	if model == "all" || model == "client" {
		genners.AddGen(types.ScopeService, ServiceClientTaskGen)
	}

	return genners
}

var ServiceBaseTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {

	tasks := make([]*types.Task, 0)

	err := geni.CheckServiceScop()
	if err != nil {
		return nil, err
	}

	// _pkgInfo := geni.PkgInfo
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

	return tasks, nil
}

var ServiceServerTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
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

	// serverfile FIXME 属于适配层的
	svrName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, "_server.go")
	svrTask := &types.Task{
		Name:      svrName,
		Path:      path.Join(outpath, svrName),
		Text:      tpl.ServerTpl,
		SetImport: tpl.ServerImportsSetter,
	}
	tasks = append(tasks, svrTask)

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

// ServiceClientTaskGen 服务级别客户端任务
var ServiceClientTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	err := geni.CheckServiceScop()
	if err != nil {
		return nil, err
	}

	_svcInfo := geni.ServiceInfo

	lowerSvcName := strings.ToLower(_svcInfo.ServiceName)
	outpath := path.Join("internal", "adpgen", "kitex", lowerSvcName)

	// clientfile
	cliName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, "_client.go")
	cliTask := &types.Task{
		Name:      cliName,
		Path:      path.Join(outpath, cliName),
		Text:      tpl.ClientTpl,
		SetImport: tpl.ClientImportsSetter,
	}
	tasks = append(tasks, cliTask)

	// agclientfile
	agcliName := fmt.Sprintf("%s_%s%s", "agkitex", lowerSvcName, "_agclient.go")
	agcliTask := &types.Task{
		Name:      agcliName,
		Path:      path.Join(outpath, agcliName),
		Text:      tpl.AgClientTpl,
		SetImport: tpl.AgClientImportsSetter,
	}
	tasks = append(tasks, agcliTask)

	return tasks, nil
}
