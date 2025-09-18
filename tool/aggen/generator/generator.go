package generator

import (
	"ag-core/tool/aggen/types"
	"fmt"
	"log/slog"
	"path/filepath"
)

var fs []*types.File

type TaskGenFunc func(*types.GennerInfo) ([]*types.Task, error)

// TaskGenerators 生成器集合
type TaskGenerators struct {
	// Gens 各阶段生成器集合 key:生成级别 value:生成器集合
	Gens map[types.ScopeType][]TaskGenFunc
}

func (tg *TaskGenerators) AddGen(scope types.ScopeType, gen TaskGenFunc) {

	if tg.Gens == nil {
		tg.Gens = make(map[types.ScopeType][]TaskGenFunc)
	}

	if _, ok := tg.Gens[scope]; !ok {
		tg.Gens[scope] = make([]TaskGenFunc, 0)
	}

	tg.Gens[scope] = append(tg.Gens[scope], gen)
}

// // ITaskGenerator task生成器接口
// type ITaskGenerator interface {
// 	GenTasks(*types.GennerInfo) ([]*types.Task, error)
// }

// // SimpleTaskGenerator 生成器
// type SimpleTaskGenerator struct {
// 	// GenTask 生成task
// 	genTasks func(*types.GennerInfo) ([]*types.Task, error)
// }

// // NewGenner 创建生成器
// func NewGenner(genTasksFunc func(*types.GennerInfo) ([]*types.Task, error)) *SimpleTaskGenerator {
// 	return &SimpleTaskGenerator{
// 		genTasks: genTasksFunc,
// 	}
// }

// // GenTasks 实现IGenner
// func (g *SimpleTaskGenerator) GenTasks(geni *types.GennerInfo) ([]*types.Task, error) {
// 	if g.genTasks == nil {
// 		return nil, fmt.Errorf("genner GenTasks func is nil")
// 	}
// 	return g.genTasks(geni)
// }

// GenRender 生成器渲染
func GenRender(geni *types.GennerInfo, taskGens *TaskGenerators) error {
	if geni == nil {
		return fmt.Errorf("geni is nil")
	}

	fs = make([]*types.File, 0)

	globalInfo := geni.GlobalInfo
	ggroups := globalInfo.PackageGroups

	// module genner
	geni.ResetSource()
	geni.GenScope = string(types.ScopeModule)
	for _, ggroup := range ggroups {
		for _, pkg := range ggroup.PackageInfos {
			// module级别的所有源文件
			geni.AddSource(pkg.IDLName)
		}
	}
	err := doGenRender(taskGens, geni, types.ScopeModule)
	if err != nil {
		return err
	}

	for _, ggroup := range ggroups {
		geni.PackageGroup = ggroup
		geni.PkgInfo = &ggroup.PkgInfo
		geni.ResetSource()
		for _, pkg := range ggroup.PackageInfos {
			// package group 的所有源文件
			geni.AddSource(pkg.IDLName)
		}
		geni.GenScope = string(types.ScopePackageGroup)

		// package group genner
		err = doGenRender(taskGens, geni, types.ScopePackageGroup)
		if err != nil {
			return err
		}

		// package genner (proto文件级别)
		for _, pkg := range ggroup.PackageInfos {
			geni.PackageInfo = pkg

			geni.ResetSource()
			geni.AddSource(pkg.IDLName)
			geni.GenScope = string(types.ScopePackage)
			err = doGenRender(taskGens, geni, types.ScopePackage)
			if err != nil {
				return err
			}

			svs := pkg.Services
			for _, sv := range svs {
				geni.ServiceInfo = sv
				// pkg.ServiceInfo = sv // 保持原逻辑，给pkg也赋值

				// service genner
				geni.GenScope = string(types.ScopeService)
				err = doGenRender(taskGens, geni, types.ScopeService)
				if err != nil {
					return err
				}
			}

		}
	}

	// 执行文件写入
	err = doWriteFile(fs)
	if err != nil {
		return err
	}

	return nil
}

// doGenRender 执行生成器渲染
func doGenRender(genners *TaskGenerators, geni *types.GennerInfo, scop types.ScopeType) error {
	gners, ok := genners.Gens[scop]
	if !ok {
		// slog.Info("genner not found", "scope", scop)
		return nil
	}

	tasks := make([]*types.Task, 0)

	for _, g := range gners {
		if g == nil {
			continue
		}

		// ts, err := g.GenTasks(geni)
		ts, err := g(geni)
		if err != nil {
			return err
		}
		tasks = append(tasks, ts...)
	}
	// TODO 是否task集中管理，此处在迭代中直接执行task的渲染
	return doTaskRender(geni, tasks)
}

// doTaskRender 执行任务渲染
func doTaskRender(geni *types.GennerInfo, tasks []*types.Task) error {
	for _, task := range tasks {
		// slog.Info(fmt.Sprintf("gen name: %-35s path: %s", task.Name, task.Path))
		slog.Info(fmt.Sprintf("gen path: %s", task.Path))

		// 清理imports
		// geni.Imports = make(map[string]map[string]bool)
		geni.ResetImport()

		// 设置imports
		if task.SetImport != nil {
			err := task.SetImport(geni)
			if err != nil {
				return err
			}
		}

		// 获取task.Path文件路径的当前目录名
		genPkgRefName := filepath.Base(filepath.Dir(task.Path))
		geni.GenPkgRefName = genPkgRefName

		f, err := task.Render(geni)
		if err != nil {
			return err
		}
		fs = append(fs, f)
	}

	return nil
}
