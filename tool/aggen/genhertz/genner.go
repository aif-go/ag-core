package genhertz

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/genhertz/tpl"
	"ag-core/tool/aggen/types"
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"strings"
)

// HertzGenServiceTask .
func HertzGenServiceTask() *generator.TaskGenerators {
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

	apipath := strings.ReplaceAll(_pkgInfo.PkgName, ".", string(filepath.Separator))
	lowerSvcName := strings.ToLower(_svcInfo.ServiceName)

	// outpath: xxx/api/hzw/hello 每个接口一个单独包目录
	outpath := path.Join(apipath, lowerSvcName)
	if !hasHttp(_svcInfo) {
		slog.Info("[genhertz] service not http method, whill skip.", "service", _svcInfo.ServiceName)
		return tasks, nil
	}

	// server
	// aghertz_xxx_server.go
	svrName := fmt.Sprintf("%s_%s_%s", "aghertz", lowerSvcName, "server.go")
	svrTask := &types.Task{
		Name:      svrName,
		Path:      path.Join(outpath, svrName),
		Text:      tpl.ServerTpl,
		SetImport: tpl.ServerImportsSetter,
	}
	tasks = append(tasks, svrTask)

	// client
	cliName := fmt.Sprintf("%s_%s_%s", "aghertz", lowerSvcName, "client.go")
	cliTask := &types.Task{
		Name:      cliName,
		Path:      path.Join(outpath, cliName),
		Text:      tpl.ClientTpl,
		SetImport: tpl.ClientImportsSetter,
	}
	tasks = append(tasks, cliTask)

	// fx
	fxName := fmt.Sprintf("%s_%s_%s", "aghertz", lowerSvcName, "fx.go")
	fxTask := &types.Task{
		Name:      fxName,
		Path:      path.Join(outpath, fxName),
		Text:      tpl.FxTpl,
		SetImport: tpl.FxImportsSetter,
	}
	tasks = append(tasks, fxTask)

	// fxadpinitfile
	fxAdpInitName := fmt.Sprintf("%s_%s_%s%s", "zfx_aghertz", _pkgInfo.PkgRefName, lowerSvcName, "_adpinit.go")
	fxAdpInitTask := &types.Task{
		Name:      fxAdpInitName,
		Path:      path.Join("api", "adpinit", fxAdpInitName),
		Text:      tpl.FxAdpInitTpl,
		SetImport: tpl.FxAdpInitImportsSetter,
	}
	tasks = append(tasks, fxAdpInitTask)

	return tasks, nil
}

func hasHttp(svc *types.ServiceInfo) bool {

	for _, mi := range svc.Methods {
		if len(mi.HttpDescs) > 0 {
			return true
		}
	}
	return false
}
