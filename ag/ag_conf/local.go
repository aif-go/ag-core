package ag_conf

import (
	"ag-core/ag/ag_conf/reader"
	"ag-core/ag/ag_ext"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const (
	AppConfKey = "app.conf"
)

type LocalConfigLoded string

var localConfLoadOnce sync.Once

// LoadLocalConfigToState 加载本地配置并返回一个标志，用于其他组件控制初始化顺序
func LoadLocalConfigToState(env IConfigurableEnvironment) (LocalConfigLoded, error) {
	err := LoadLocalConfig(env)
	return LocalConfigLoded("localConfigLoaded"), err
}

// LoadLocalConfig 加载本地配置 本地支持yaml|yml|properties后缀的三个文件
// 1. 先判断环境变量或者进程变量中是否配置app.conf对应的key的值,未配置取执行二进制文件中的app.suffix文件
// 2. 获取app.profile的环境的设置,例如dev,sit,uat等
// 3. 如果1的场景未配置，则按照app.suffix app_profile.suffix的顺序加载,后续的内容会覆盖前者
func LoadLocalConfig(env IConfigurableEnvironment) (rerr error) {
	localConfLoadOnce.Do( // TODO 能否切换幂等逻辑，能重复加载本地配置，非全局的控制，以适应不同场景的重复加载
		func() {
			rerr = LoadLocalConfigRepeatable(env)
		},
	)
	return
}

func LoadLocalConfigRepeatableToState(env IConfigurableEnvironment) (LocalConfigLoded, error) {
	err := LoadLocalConfigRepeatable(env)
	return LocalConfigLoded("localConfigRepeatableLoaded"), err
}

func LoadLocalConfigRepeatable(env IConfigurableEnvironment) (rerr error) {
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
	return
}

func doLoadLocalConfig(env IConfigurableEnvironment) error {

	slog.Info("--- LoadLocalConfig ---")
	appConf := env.GetProperty(AppConfKey)
	if appConf == "" {
		// 获取当前main.go的目录
		// dir, err := os.Getwd()
		dir, err := os.Executable()
		if err != nil {
			return err
		}
		dir = filepath.Dir(dir)
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

	// 获取文件的路径
	appConfFile, err := filepath.Abs(appConf)
	if err != nil {
		return err
	}
	slog.Info(fmt.Sprintf("config file: %s", appConfFile))

	return LoadConfigFile(env, appConfFile)

}

func LoadConfigFile(env IConfigurableEnvironment, appConfFile string) error {
	// context, err := os.ReadFile(appConfFile)
	// if err != nil {
	// 	return err
	// }
	// // suffix := filepath.Ext(appConfFile)
	// acftype := format(appConfFile)
	// reader, ok := reader.Readers[acftype]
	// if !ok {
	// 	return fmt.Errorf("app.conf type not supported: %s", acftype)
	// }

	// contextMap, err := reader(context)
	// if err != nil {
	// 	return err
	// }

	// flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	// if err != nil {
	// 	return err
	// }

	// env.GetPropertySources().AddLast(&MapPropertySource{ // TODO 本地配置文件是否应该比SYS优先级高？ FIXME 应该SYS < localfile < nacos < -D
	// 	NamedPropertySource: NamedPropertySource{
	// 		Name: fmt.Sprintf("%s-%s", SourceKeyLocalPrefix, appConfFile), // "[LOCAL]-xxxx"
	// 	},
	// 	Source: flatmapcontext,
	// })

	propertySource, err := NewPropertySourceFromFile(appConfFile)
	if err != nil {
		return err
	}
	env.GetPropertySources().AddLast(propertySource)
	return nil

}

func LoadConfigDir(env IConfigurableEnvironment, appConfFile string) error {
	// TODO loadDir
	err := fmt.Errorf("app.conf is a directory")
	slog.Error("loadConfigDir", "err", err)
	return err
}

func NewPropertySourceFromFile(confFile string) (IPropertySource, error) {
	context, err := os.ReadFile(confFile)
	if err != nil {
		return nil, err
	}
	acftype := format(confFile)
	reader, ok := reader.Readers[acftype]
	if !ok {
		return nil, fmt.Errorf("config type not supported: %s", acftype)
	}

	contextMap, err := reader(context)
	if err != nil {
		return nil, err
	}

	flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	if err != nil {
		return nil, err
	}

	sps := &MapPropertySource{ // TODO 本地配置文件是否应该比SYS优先级高？ FIXME 应该SYS < localfile < nacos < -D
		NamedPropertySource: NamedPropertySource{
			Name: fmt.Sprintf("%s-%s", SourceKeyLocalPrefix, confFile), // "[LOCAL]-xxxx"
		},
		Source: flatmapcontext,
	}
	return sps, nil
}

func format(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx+1:]
	}
	return ""
}
