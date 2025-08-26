package genner

import (
	"ag-core/cmd/protoc-gen-go-agkitex/aggen"
	"ag-core/cmd/protoc-gen-go-agkitex/tpl"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

func Genner(pkg *aggen.PackageInfo) ([]*aggen.Task, error) {
	// 获取文件输出路径
	// output := pkg.ImportPath
	output := pkg.PkgName
	output = strings.ReplaceAll(output, ".", string(filepath.Separator)) // 将pkg名替换为系统文件路径
	svcPkg := strings.ToLower(pkg.ServiceName)
	output = path.Join(output, svcPkg) // package/service/

	// servicefile
	svcName := fmt.Sprintf("%s_%s%s", pkg.PluginSortName, strings.ToLower(pkg.ServiceName), ".go")
	svcTask := &aggen.Task{
		Name: svcName,
		Path: path.Join(output, svcName),
		Text: tpl.ServiceTpl,
	}

	// clientfile
	cliName := fmt.Sprintf("%s_%s", pkg.PluginSortName, "client.go")
	cliTask := &aggen.Task{
		Name: cliName,
		Path: path.Join(output, cliName),
		Text: tpl.ClientTpl,
		// SetImport: tpl.ClientImportsSetter,
		SetImport: tpl.ClientImportsSetter,
	}
	/*
		// serverfile
		svrName := fmt.Sprintf("%s_%s", pluginName, "server.go")
		svrTask := &Task{
			Name: svrName,
			Path: path.Join(output, svrName),
			Text: tpl.ServerTpl,
		}
		// fxfile
		fxName := fmt.Sprintf("%s_%s", pluginName, "fx.go")
		fxTask := &Task{
			Name: fxName,
			Path: path.Join(output, fxName),
			Text: tpl.FxTpl,
		}
	*/

	// tasks := []*Task{cliTask, svrTask, svcTask, fxTask}
	tasks := []*aggen.Task{svcTask, cliTask}

	return tasks, nil
}
