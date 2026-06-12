package genhertz

import (
	"github.com/aif-go/ag-core/tool/aggen/generator"
	"github.com/aif-go/ag-core/tool/aggen/genhertz/tpl"
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
	"log/slog"
	"path"
	"strings"
)

// HertzGenServiceTask .
func HertzGenServiceTask(model string) *generator.TaskGenerators {
	genners := &generator.TaskGenerators{}

	if model == "all" || model == "client" {
		genners.AddGen(types.ScopeService, ServiceClientTaskGen)
	}

	if model == "all" || model == "server" {
		genners.AddGen(types.ScopeService, ServiceServerTaskGen)
	}

	return genners
}

var ServiceClientTaskGen = func(geni *types.GennerInfo) ([]*types.Task, error) {
	tasks := make([]*types.Task, 0)

	err := geni.CheckServiceScop()
	if err != nil {
		return nil, err
	}

	// _pkgInfo := geni.PkgInfo
	_svcInfo := geni.ServiceInfo

	// apipath := strings.ReplaceAll(_pkgInfo.PkgName, ".", string(filepath.Separator))
	lowerSvcName := strings.ToLower(_svcInfo.ServiceName)

	// outpath: xxx/api/hzw/hello 每个接口一个单独包目录
	// outpath := path.Join(apipath, lowerSvcName)
	outpath := path.Join("internal", "adpgen", "hertz", lowerSvcName)

	if !hasHttp(_svcInfo) {
		slog.Info("[genhertz] service not http method, whill skip.", "service", _svcInfo.ServiceName)
		return tasks, nil
	}

	// client
	cliName := fmt.Sprintf("%s_%s_%s", "aghertz", lowerSvcName, "client.go")
	cliTask := &types.Task{
		Name: cliName,
		Path: path.Join(outpath, cliName),
		// Text:      tpl.ClientTpl,
		// SetImport: tpl.ClientImportsSetter,
		Text:      tpl.ClientTplV2,
		SetImport: tpl.ClientImportsSetterV2,
	}
	tasks = append(tasks, cliTask)

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

	// outpath: xxx/api/hzw/hello 每个接口一个单独包目录
	// outpath := path.Join(apipath, lowerSvcName)
	outpath := path.Join("internal", "adpgen", "hertz", lowerSvcName)

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
	adpPath := "internal/adpgen"
	adpInitPath := path.Join(adpPath, "adpinit")

	fxAdpInitName := fmt.Sprintf("%s_%s_%s%s", "zfx_aghertz", _pkgInfo.PkgRefName, lowerSvcName, "_adpinit.go")
	fxAdpInitTask := &types.Task{
		Name: fxAdpInitName,
		// Path:      path.Join("api", "adpinit", fxAdpInitName),
		Path:      path.Join(adpInitPath, fxAdpInitName),
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
