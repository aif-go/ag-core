package cmd_proto

import (
	"os"
	"path/filepath"
	"strings"
)

// findProtos 查找所有的proto文件
func findProtos(fs []string) ([]string, error) {
	var protos []string
	for _, f := range fs {
		ps, err := doFindProtos(f)
		if err != nil {
			return nil, err
		}
		protos = append(protos, ps...)
	}
	return protos, nil
}

func doFindProtos(f string) ([]string, error) {
	var protos []string

	// f is dir
	fi, err := os.Stat(f)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		// 遍历目录
		fds, err := os.ReadDir(f)
		if err != nil {
			return nil, err
		}
		for _, fd := range fds {
			fp := filepath.Join(f, fd.Name())
			_ps, err := doFindProtos(fp)
			if err != nil {
				return nil, err
			}
			protos = append(protos, _ps...)
		}
	} else {
		if strings.HasSuffix(f, ".proto") {
			protos = append(protos, f)
		}
	}

	return protos, nil
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}
