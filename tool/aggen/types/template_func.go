package types

import (
	"strings"
	"time"

	"github.com/cloudwego/kitex/tool/internal_pkg/util"
)

// funcs 模板函数，提供给模板调用的函数工具，调用模板时会将该集合添加到模板执行中
var funcs = map[string]interface{}{
	"ToLower":                 strings.ToLower,
	"LowerFirst":              util.LowerFirst,
	"UpperFirst":              util.UpperFirst,
	"NotPtr":                  util.NotPtr,
	"ReplaceString":           util.ReplaceString,
	"SnakeString":             util.SnakeString,
	"backquoted":              BackQuoted,
	"DateTimeNow":             func() string { return time.Now().Format(time.DateTime) },
	"ParamUnptrTypeInPkgInfo": ParameterUnptrTypeInPkgInfo,
}

// AddTemplateFunc 添加模板函数
func AddTemplateFunc(key string, f interface{}) {
	funcs[key] = f
}

func BackQuoted(s string) string {
	return "`" + s + "`"
}

// ParameterUnptrTypeInPkgInfo 从类型中移除包前缀，同包下引用无需前缀，否则循环依赖
func ParameterUnptrTypeInPkgInfo(param Parameter, pkgInfo PkgInfo) string {
	pkgRefName := pkgInfo.PkgRefName
	pupType := param.UnptrType
	// pupType 若是 'pkgRefName.’ 开头，则移除
	if strings.HasPrefix(pupType, pkgRefName+".") {
		pupType = strings.TrimPrefix(pupType, pkgRefName+".")
	}

	return pupType
}
