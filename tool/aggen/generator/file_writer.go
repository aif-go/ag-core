package generator

import (
	"github.com/aif-go/ag-core/tool/aggen/types"
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func doWriteFile(fs []*types.File) error {
	// 检查文件情况，如重复覆盖等情况
	err := fscheck(fs)
	if err != nil {
		return err
	}

	// 执行文件写入
	for _, f := range fs {
		err := tryWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

func tryWriteFile(f *types.File) error {
	switch f.OutputType {
	case types.OutputTypeOverride:
		// 覆盖文件
		return writeFile(f.Name, []byte(f.Content))
	case types.OutputTypeAppend:
		return fmt.Errorf("append not implement")
	case types.OutputTypeInsert:
		return fmt.Errorf("insert not implement")
	case types.OutputTypeNotOverride:
		_, err := os.Stat(f.Name)
		if err == nil {
			slog.Warn(fmt.Sprintf("file %s already exist, skip", f.Name))
			return nil
		}
		return writeFile(f.Name, []byte(f.Content))
	default:
		// 其他类型，默认覆盖
		return writeFile(f.Name, []byte(f.Content))
	}

	return nil
}

func writeFile(fn string, data []byte) error {
	if strings.HasSuffix(fn, ".go") {
		// 格式化代码
		formatted, err := format.Source(data)
		if err != nil {
			slog.Error("format file err", "file", fn, "err", err)
			// return fmt.Errorf("format file %s err: %s", fn, err)
			// log.Errorf("format file %q err: %s", fn, err)
		} else {
			data = formatted
		}
	}
	err := os.MkdirAll(filepath.Dir(fn), 0o755)
	if err == nil {
		err = os.WriteFile(fn, data, 0o644)
	}
	if err != nil {
		return fmt.Errorf("write file %q err: %s", fn, err)
		// log.Errorf("write file %q err: %s", fn, err)
	}
	return nil
}

func fscheck(fs []*types.File) error {
	// 检查文件是否有覆盖重复
	overrideFiles := make(map[string]int)
	for _, f := range fs {
		if f.OutputType == types.OutputTypeOverride {
			// 检查是否有重复
			if _, ok := overrideFiles[f.Name]; ok {
				return fmt.Errorf("file %s override repeat", f.Name)
			}
			overrideFiles[f.Name] = 1
		}
	}

	// FIXME TODO 交互式确认文件覆盖

	return nil
}
