package util

import (
	"github.com/aif-go/ag-core/tool/aggen/types"
	"encoding/json"
	"fmt"
	"go/token"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"

	"github.com/davecgh/go-spew/spew"
)

// GoSanitized returns a valid Go identifier. 字符合法化
func GoSanitized(s string) string {
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

// FixImport 清理import中的引号
func FixImport(path string) string {
	path = strings.Trim(path, "\"")
	return path
}

// GetPwdModInfo 解析当前目录下的gomod信息
func GetPwdModInfo() (*types.ModuleInfo, error) {
	curpath, err := filepath.Abs(".")
	if err != nil {
		return &types.ModuleInfo{}, err
	}

	modName, modPath, hasMod := SearchGoMod(curpath)

	mi := &types.ModuleInfo{
		PwdGoMod:     modName,
		PwdGoModPath: modPath,
		HasPwdGoMod:  hasMod,
	}

	return mi, nil
}

// SearchGoMod searches go.mod from the given directory (which must be an absolute path) to
// the root directory. When the go.mod is found, its module name and path will be returned.
func SearchGoMod(cwd string) (moduleName, path string, found bool) {
	for {
		path = filepath.Join(cwd, "go.mod")
		data, err := os.ReadFile(path)
		if err == nil {
			re := regexp.MustCompile(`^\s*module\s+(\S+)\s*`)
			for _, line := range strings.Split(string(data), "\n") {
				m := re.FindStringSubmatch(line)
				if m != nil {
					return m[1], cwd, true
				}
			}
			return fmt.Sprintf("<module name not found in '%s'>", path), path, true
		}

		if !os.IsNotExist(err) {
			return
		}
		parentCwd := filepath.Dir(cwd)
		if parentCwd == cwd {
			break
		}
		cwd = parentCwd
	}
	return
}

var config = spew.ConfigState{
	MaxDepth: 15, // 限制递归深度
	Indent:   "  ",
}

func TestLog(msg string, obj any) {

	// if !slog.Default().Enabled(nil, slog.LevelDebug) {
	// 	return
	// }

	pijson, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		msgstring := config.Sdump(obj)
		slog.Info(fmt.Sprintf("%s:%s", msg, msgstring))
	} else {
		slog.Info(fmt.Sprintf("%s:%s", msg, pijson))
	}
}
