package types

import (
	"errors"
	"maps"
)

// GennerInfo 提供给模板最终使用的对象，各生成层级会将对应层级数据提取平铺到该实体中
type GennerInfo struct {
	// 插件名 (主插件名，生成文件的主体)
	pluginName string

	// TODO 此处是否需要添加插件版本信息，若后续调整为集中生成，则这个版本信息怎么设置
	Versions     map[string]string // 版本信息： key:插件名，value:插件版本。 用于模板中显示生成的插件信息
	BaseVersions map[string]string // 基础的版本信息
	ModuleInfo   *ModuleInfo       // 当前模块的信息

	GlobalInfo *GlobalInfo // 全局信息，包含所有信息，下面各阶段缓存信息都从全局信息中提取

	PkgInfo       *PkgInfo                   // (缓存) 包元信息(缓存，便于模板取用)
	PackageGroup  *PackageGroup              // (缓存) 包组信息(缓存,便于模板取用)
	PackageInfo   *PackageInfo               // (缓存) 包信息(IDL文件级别)(缓存,便于模板取用)
	ServiceInfo   *ServiceInfo               // (缓存) 服务信息(缓存，便于模板取用)
	Imports       map[string]map[string]bool // (缓存) import path => alias
	Source        []string                   // (缓存) 源文件
	GenPkgRefName string                     // (缓存) 生成的包名(引用名) 来源于task的文件实际生成路径，主要用来生成文件头package部分
	GenScope      string                     // (缓存) 生成级别
}

// Reset 重置GennerInfo，用于下一次生成
func (gi *GennerInfo) Reset() {
	gi.Imports = make(map[string]map[string]bool)
	// gi.GlobalInfo = nil // 全局信息，不重置
	gi.PackageGroup = nil
	gi.PackageInfo = nil
	gi.ServiceInfo = nil
	gi.Source = make([]string, 0)
}

// SetVersion 设置相关版本信息，一般是对应插件的信息，用于在模板生成时显示相关插件版本
func (gi *GennerInfo) SetBaseVersion(pluginName, version string) {
	if gi.BaseVersions == nil {
		gi.BaseVersions = make(map[string]string)
	}
	if gi.Versions == nil {
		gi.Versions = maps.Clone(gi.BaseVersions)
	}
	gi.BaseVersions[pluginName] = version

	maps.Copy(gi.Versions, gi.BaseVersions)

}

// ResetBaseVersions 重置基础版本信息
func (gi *GennerInfo) ResetBaseVersions() {
	if len(gi.Versions) > 0 {
		// 移除Versions中BaseVersion的key
		for k := range gi.Versions {
			if _, ok := gi.BaseVersions[k]; ok {
				delete(gi.Versions, k)
			}
		}
	}
	gi.BaseVersions = make(map[string]string)
}

// SetVersion 设置相关版本信息，一般是对应插件的信息，用于在模板生成时显示相关插件版本
func (gi *GennerInfo) SetVersion(pluginName, version string) {
	if gi.BaseVersions == nil {
		gi.BaseVersions = make(map[string]string)
	}
	if len(gi.Versions) == 0 {
		gi.Versions = maps.Clone(gi.BaseVersions)
	}

	gi.Versions[pluginName] = version
}

// ResetVersion .
func (gi *GennerInfo) ResetVersion() {
	gi.Versions = maps.Clone(gi.BaseVersions)
}

// AddPackageInfo 添加packageInfo
func (gi *GennerInfo) AddPackageInfo(pkg *PackageInfo) error {
	if gi.GlobalInfo == nil { // FIXME 注意：当前实现线程不安全，若后续有并发场景，需要加锁
		gi.GlobalInfo = &GlobalInfo{}
	}

	err := gi.GlobalInfo.AddPackageInfo(pkg) // 约束检查可能产生异常
	if err != nil {
		return err
	}

	return nil
}

// SetPluginName 设置主插件名
func (gi *GennerInfo) SetPluginName(pluginName string) {
	gi.pluginName = pluginName
}

// PluginName 获取插件名
func (gi *GennerInfo) PluginName() string {
	if gi.pluginName == "" {
		return "agcore_gen"
	}
	return gi.pluginName
}

// ResetSource 重置源文件
func (g *GennerInfo) ResetSource() {
	g.Source = make([]string, 0)
}

// AddSource 添加源文件
func (g *GennerInfo) AddSource(source string) {
	g.Source = append(g.Source, source)
}

// ResetImport .
func (g *GennerInfo) ResetImport() {
	g.Imports = make(map[string]map[string]bool)
}

// AddImport .
func (g *GennerInfo) AddImport(pkg, path string) {
	if g.Imports == nil {
		g.Imports = make(map[string]map[string]bool)
	}
	if pkg != "" {
		// path = g.toExternalGenPath(path)
		if path == pkg {
			g.Imports[path] = nil
		} else {
			if g.Imports[path] == nil {
				g.Imports[path] = make(map[string]bool)
			}
			g.Imports[path][pkg] = true
		}
	}
}

// AddImports .
func (g *GennerInfo) AddImports(pkgs ...string) {
	for _, pkg := range pkgs {
		if path, ok := globalDependencies[pkg]; ok {
			g.AddImport(pkg, path)
		} else {
			g.AddImport(pkg, pkg)
		}
	}
}

// CheckModuleScop .
func (g *GennerInfo) CheckModuleScop() error {
	if g.ModuleInfo == nil {
		return errors.New("ModuleInfo is nil")
	}
	if g.GlobalInfo == nil {
		return errors.New("GlobalInfo is nil")
	}
	return nil
}

// CheckPackageGroupScop .
func (g *GennerInfo) CheckPackageGroupScop() error {
	if g.PkgInfo == nil {
		return errors.New("PkgInfo is nil")
	}
	if g.PackageGroup == nil {
		return errors.New("PackageGroup is nil")
	}

	return g.CheckModuleScop()
}

// CheckPackageScop .
func (g *GennerInfo) CheckPackageScop() error {
	if g.PackageInfo == nil {
		return errors.New("PackageInfo is nil")
	}
	return g.CheckPackageGroupScop()
}

// CheckServiceScop .
func (g *GennerInfo) CheckServiceScop() error {
	if g.ServiceInfo == nil {
		return errors.New("ServiceInfo is nil")
	}
	return g.CheckPackageScop()
}
