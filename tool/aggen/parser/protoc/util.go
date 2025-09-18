package protoc

import (
	"fmt"
	"go/token"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
	"google.golang.org/protobuf/compiler/protogen"
)

// protocVersion returns the version of the protoc compiler.
func protocVersion(gen *protogen.Plugin) string {
	v := gen.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

// goSanitized returns a valid Go identifier. 字符合法化
func goSanitized(s string) string {
	// replace invalid characters with '_'
	s = strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return '_'
		}
		return r
	}, s)

	// avoid invalid identifier
	if !token.IsIdentifier(s) {
		return "_" + s
	}
	return s
}

var config = spew.ConfigState{
	MaxDepth: 15, // 限制递归深度
	Indent:   "  ",
}

func testLog(msg string, obj any) {

	// if !slog.Default().Enabled(nil, slog.LevelDebug) {
	// 	return
	// }

	// pijson, err := json.MarshalIndent(obj, "", "  ")
	// if err != nil {
	// 	msgstring := config.Sdump(obj)
	// 	slog.Error(fmt.Sprintf("%s:%s", msg, msgstring))
	// } else {
	// 	slog.Error(fmt.Sprintf("%s:%s", msg, pijson))
	// }
}
