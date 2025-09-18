package types

import "fmt"

// GlobalInfo 为顶层数据，一些全局源码文件的生成以此为数据基础
type GlobalInfo struct {
	PackageGroups map[string]*PackageGroup // 包信息，key:包名.
	ModuleInfo    *ModuleInfo              // 当前模块的信息，此为冗余，便于模板中取用
}

// ModuleInfo 当前模块的信息
type ModuleInfo struct {
	PwdGoMod     string // 当前模块的名称
	PwdGoModPath string // 当前模块的路径
	HasPwdGoMod  bool   // 当前模块是否存在go.mod文件
}

// PackageGroup 包组，每个proto文件对应一个PackageInfo，可能多个proto文件属于同一个pkg
type PackageGroup struct {
	PkgInfo      PkgInfo        // 约束：所有PackageInfo的 PkgName 和 ImportPath 必须相同，否则解析应该报错
	PackageInfos []*PackageInfo // 每个proto文件对应一个PackageInfo，相同包名的pki在同一个切片中
	ModuleInfo   *ModuleInfo    // 当前模块的信息，此为冗余，便于模板中取用
}

// AddPackageInfo 添加PackageInfo
func (ggi *GlobalInfo) AddPackageInfo(pkg *PackageInfo) error {
	if ggi.PackageGroups == nil {
		ggi.PackageGroups = make(map[string]*PackageGroup)
	}
	pkgName := pkg.PkgInfo.PkgName
	pkgGroup, ok := ggi.PackageGroups[pkgName]
	if !ok {
		pkgGroup = &PackageGroup{
			PkgInfo:      pkg.PkgInfo,
			PackageInfos: []*PackageInfo{pkg},
			ModuleInfo:   ggi.ModuleInfo,
		}
		ggi.PackageGroups[pkgName] = pkgGroup
	} else {
		// 约束：所有PackageInfo的 PkgName 和 ImportPath 必须相同，否则解析应该报错
		if pkg.PkgInfo.ImportPath != pkgGroup.PkgInfo.ImportPath {
			return fmt.Errorf("package %s import path not same", pkg.PkgInfo.PkgName)
		}
		pkgGroup.PackageInfos = append(pkgGroup.PackageInfos, pkg)
	}
	return nil
}
