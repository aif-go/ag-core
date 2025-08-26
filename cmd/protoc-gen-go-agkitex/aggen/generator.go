package aggen

import (
	"fmt"
	"go/format"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/kitex/tool/internal_pkg/log"
)

// GenTasks 生成任务函数，该函数执行级别为services
type GenTasks func(pkg *PackageInfo) ([]*Task, error)

// GenPkg 包级别的代码生成，每个包会执行一遍GenTasks
func GenPkg(p *PackageInfo, gtf GenTasks) error {
	fs, err := generateService(p, gtf)
	if err != nil {
		return err
	}
	for _, f := range fs {
		writeFile(f.Name, []byte(f.Content))
	}
	return nil
}

// GenService Service级别的代码生成，每个service会执行一遍GenTasks
func GenService(p *PackageInfo, gtf GenTasks) error {
	for _, s := range p.Services {
		p.ServiceInfo = s // 设置 target service

		// TODO 现应该可以扩展
		fs, err := generateService(p, gtf)
		if err != nil {
			return err
		}
		for _, f := range fs {
			writeFile(f.Name, []byte(f.Content))
		}
	}

	return nil
}

func generateService(pkg *PackageInfo, gtf GenTasks) ([]*File, error) {
	var fs []*File
	// slog.Info("generateServiceFiles:", "PkgName", pkg.PkgName)
	// slog.Info("generateServiceFiles:", "PkgRefName", pkg.PkgRefName)
	// slog.Info("generateServiceFiles:", "ImportPath", pkg.ImportPath)
	// slog.Info("generateServiceFiles:", "ServiceName", pkg.ServiceName)
	// slog.Info("generateServiceFiles:", "RawServiceName", pkg.RawServiceName)

	// // 获取文件输出路径
	// // output := pkg.ImportPath
	// output := pkg.PkgName
	// output = strings.ReplaceAll(output, ".", string(filepath.Separator)) // 将pkg名替换为系统文件路径
	// svcPkg := strings.ToLower(pkg.ServiceName)
	// output = path.Join(output, svcPkg) // package/service/

	// // 获取绝对路径
	// absPath, err := filepath.Abs(output)
	// if err != nil {
	// 	fmt.Printf("获取绝对路径失败: %v\n", err)
	// 	return nil, err
	// }

	// slog.Info("generateServiceFiles:", "output", output)
	// slog.Info("generateServiceFiles:", "absPath", absPath) // TODO 是否使用绝对路径
	// slog.Info(fmt.Sprintf("generateServiceFiles: IDLName:%s", pkg.IDLName))

	tasks, err := gtf(pkg)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		slog.Info(fmt.Sprintf("generateServiceFiles: name:%-35s path:%s", task.Name, task.Path))

		// 清理imports
		pkg.Imports = make(map[string]map[string]bool)
		// 设置imports
		if task.SetImport != nil {
			task.SetImport(pkg)
		}

		f, err := task.Render(pkg)
		if err != nil {
			return nil, err
		} else {
			fs = append(fs, f)
		}
	}
	return fs, nil
}

func writeFile(fn string, data []byte) {
	if strings.HasSuffix(fn, ".go") {
		// 格式化代码
		formatted, err := format.Source(data)
		if err != nil {
			log.Errorf("format file %q err: %s", fn, err)
		} else {
			data = formatted
		}
	}
	err := os.MkdirAll(filepath.Dir(fn), 0o755)
	if err == nil {
		err = os.WriteFile(fn, data, 0o644)
	}
	if err != nil {
		log.Errorf("write file %q err: %s", fn, err)
	}
}
