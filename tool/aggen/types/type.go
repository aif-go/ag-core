// copyright 2023 cloudwego
// ext by houzw
package types

// PkgInfo .
type PkgInfo struct {
	PkgName    string // 来源 PROTO：package eg: api.hello
	ImportPkg  string // PkgName 的导入形态， 例：proto 中 pacakge 内容'.'替换为'/' eg: api/hello . add by houzw
	PkgRefName string // 来源 PROTO：option go_package, path.Base(ImportPath) eg: hello
	ImportPath string // 来源 PROTO：option go_package 原始值 eg: xxxx/api/hello
}

// Parameter .
type Parameter struct {
	// PkgInfo的含义是不一样的哦
	// proto 中此处只有一个元素表示当前参数的import信息
	//   PkgRefName为import名，path.Base(importPath) eg: hzw
	//   ImportPath为import路径 eg: github.hzw.cn/api/hzw
	//     来源于proto文件定义：`option go_package`  TODO 可能和当前项目下不一致,也有可能和proto中的package配置不一致，具体依赖需要如何处理
	Deps []PkgInfo

	Name    string // 变量名:   目前只有 Req/Resp
	RawName string // 原始定义: 目前只有 Req/Resp
	Type    string // 带报名的指针类型名 *pkgRefName + "." + RawType, eg: *PkgA.StructB

	// ext by houzw
	UnptrType string // 非指针类型的Type eg: PkgA.StructB
	RawType   string // 不带报名的类型名 msg.GoIdent.GoName eg: StructB
}

// CommentSet 注释信息.
type CommentSet struct {
	LeadingDetacheds []string
	Leading          string
	Trailing         string
}
