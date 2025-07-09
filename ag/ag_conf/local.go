package ag_conf

import (
	"ag-core/ag/ag_conf/reader"
	"ag-core/ag/ag_ext"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type LocalConfigLoded string

var localConfLoadOnce sync.Once

// LoadLocalConfigToState 加载本地配置并返回一个标志，用于其他组件控制初始化顺序
func LoadLocalConfigToState(env IConfigurableEnvironment, localEmbed embed.FS) (LocalConfigLoded, error) {
	err := LoadLocalConfig(env, localEmbed)
	return LocalConfigLoded("localConfigLoaded"), err
}

// LoadLocalConfig 加载本地配置 本地支持yaml|yml|properties后缀的三个文件
// 1. 先判断环境变量或者进程变量中是否配置app.conf对应的key的值,未配置取执行二进制文件中的app.suffix文件
// 2. 获取app.profile的环境的设置,例如dev,sit,uat等
// 3. 如果1的场景未配置，则按照app.suffix app_profile.suffix的顺序加载,后续的内容会覆盖前者
func LoadLocalConfig(env IConfigurableEnvironment, localEmbed embed.FS) (rerr error) {
	localConfLoadOnce.Do(
		func() {
			// 加载本地配置文件
			err := doLoadLocalConfig(env)
			if err != nil {
				rerr = err
				return
			}
			// 解密LocalConfig
			err = DecryptLocalConfig(env)
			if err != nil {
				rerr = err
				return
			}
		},
	)
	return
}

func doLoadLocalConfig(env IConfigurableEnvironment) error {

	slog.Info("--- LoadLocalConfig ---")
	appConf := env.GetProperty("app.conf")
	if appConf == "" {
		// 获取当前main.go的目录
		dir, err := os.Getwd()
		if err != nil {
			return err
		}
		appConf = filepath.Join(dir, "app.yml")
		// slog.Warn("app.conf not found, will use default value:", appConf)
		slog.Warn(fmt.Sprintf("app.conf not found, will use default value: %s", appConf))
	}

	fi, err := os.Stat(appConf)
	if err != nil {
		return err
	}

	if fi.IsDir() {
		return LoadConfigDir(env, appConf)
	}

	return LoadConfigFile2(env, appConf)

}

func LoadConfigFile2(env IConfigurableEnvironment, appConfFile string) error {
	context, err := os.ReadFile(appConfFile)
	if err != nil {
		return err
	}
	// suffix := filepath.Ext(appConfFile)
	acftype := format(appConfFile)
	reader, ok := reader.Readers[acftype]
	if !ok {
		return fmt.Errorf("app.conf type not supported: %s", acftype)
	}

	contextMap, err := reader(context)
	if err != nil {
		return err
	}

	flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	if err != nil {
		return err
	}

	env.GetPropertySources().AddLast(&MapPropertySource{
		NamedPropertySource: NamedPropertySource{
			Name: fmt.Sprintf("%s-%s", SourceKeyLocalPrefix, appConfFile), // "[LOCAL]-xxxx"
		},
		Source: flatmapcontext,
	})
	return nil

}

func LoadConfigDir(env IConfigurableEnvironment, appConfFile string) error {
	// TODO loadDir
	return fmt.Errorf("app.conf is a directory")
}

func format(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return ""
}
