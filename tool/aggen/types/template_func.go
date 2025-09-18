package types

import (
	"strings"
	"time"

	"github.com/cloudwego/kitex/tool/internal_pkg/util"
)

// funcs 模板函数，提供给模板调用的函数工具，调用模板时会将该集合添加到模板执行中
var funcs = map[string]interface{}{
	"ToLower":       strings.ToLower,
	"LowerFirst":    util.LowerFirst,
	"UpperFirst":    util.UpperFirst,
	"NotPtr":        util.NotPtr,
	"ReplaceString": util.ReplaceString,
	"SnakeString":   util.SnakeString,
	"backquoted":    BackQuoted,
	"DateTimeNow":   func() string { return time.Now().Format(time.DateTime) },
}

// AddTemplateFunc 添加模板函数
func AddTemplateFunc(key string, f interface{}) {
	funcs[key] = f
}

func BackQuoted(s string) string {
	return "`" + s + "`"
}
