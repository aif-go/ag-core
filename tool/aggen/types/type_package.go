package types

// PackageInfo contains information to generate a package for a service.
type PackageInfo struct {
	// Dependencies map[string]string // 依赖包信息，用于AddImports 搜索预定的依赖信息 TODO 依赖暂全以addimport方式添加
	// *ServiceInfo                // 目标服务信息，在service级别的生成任务中，会提前从Services中提取到此处，便于模板获取信息
	Services []*ServiceInfo // 所有服务信息，在package级别的生成任务中，会提前从Services中提取到此处，便于模板获取信息

	Codec string // IDL模板类型 proto,thrift TODO 此属性应该属于idl文件级别
	// Version string // 生成插件的版本
	// Imports map[string]map[string]bool // import path => alias
	IDLName string // IDL文件名 TODO

	// -- the following ext by houzw
	Deprected  bool        // 是否废弃
	PkgInfo    PkgInfo     // 包信息，原kitex中该信息只存在于ServiceInfo中，但我们当前代码生成包含pkg层的生成，所以此处冗余作为pkg层的信息
	ModuleInfo *ModuleInfo // 当前模块的信息，此为冗余，便于模板中取用

	NoFastAPI      bool // [保留] kitex保留，仅兼容原kitex的生成模板
	FrugalPretouch bool // [保留] kitex保留，仅兼容原kitex的生成模板
	StreamX        bool // [保留] kitex保留，仅兼容原kitex的生成模板

}

// AddImport .
//func (p *PackageInfo) AddImport(pkg, path string) {
//	if p.Imports == nil {
//		p.Imports = make(map[string]map[string]bool)
//	}
//	if pkg != "" {
//		// path = p.toExternalGenPath(path) //  FIXME 将内部路径转换为外部路径，目前不需要此逻辑，后续若有需求再行评估
//		if path == pkg {
//			p.Imports[path] = nil
//		} else {
//			if p.Imports[path] == nil {
//				p.Imports[path] = make(map[string]bool)
//			}
//			p.Imports[path][pkg] = true
//		}
//	}
//}
