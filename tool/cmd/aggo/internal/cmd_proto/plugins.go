package cmd_proto

import (
	"fmt"

	"github.com/samber/lo"
)

const (
	ModelBase   = "base"
	ModelAll    = "all"
	ModelServer = "server"
	ModelClient = "client"

	PluginBase = "base"
	PluginAll  = "all"
)

var protoPluginsStor = map[string]map[string][]string{}

func RegPlugin(plugin, model string, pluginOpt string) {
	if _, ok := protoPluginsStor[plugin]; !ok {
		protoPluginsStor[plugin] = map[string][]string{}
	}

	if _, ok := protoPluginsStor[plugin][model]; !ok {
		protoPluginsStor[plugin][model] = []string{}
	}

	protoPluginsStor[plugin][model] = append(protoPluginsStor[plugin][model], pluginOpt)
}

func GetAllPluginsName() []string {
	plugins := map[string]bool{}
	for k := range protoPluginsStor {
		plugins[k] = true
	}
	return lo.Keys(plugins)
}

// selectPlugins select plugins by models
func selectPlugins(plugins []string, models []string) ([]string, error) {

	pgs := []string{}

	// 提前加载base插件
	if _, ok := protoPluginsStor[PluginBase]; ok {
		if _, ok := protoPluginsStor[PluginBase][ModelBase]; ok {
			pgs = append(pgs, protoPluginsStor[PluginBase][ModelBase]...)
		}
	}

	modelAll := lo.Contains(models, ModelAll)
	pluginAll := lo.Contains(plugins, PluginAll)

	if pluginAll {
		// 加载所有插件
		for k, mps := range protoPluginsStor {
			if k == PluginBase {
				continue
			}

			for m, _ := range mps {
				if modelAll || m == ModelBase || lo.Contains(models, m) {
					pgs = append(pgs, mps[m]...)
				}
			}
		}

	} else {
		// 加载指定插件
		for _, plugin := range plugins {
			if plugin == PluginBase {
				continue
			}
			if mps, ok := protoPluginsStor[plugin]; ok {
				for m, _ := range mps {
					if modelAll || m == ModelBase || lo.Contains(models, m) {
						pgs = append(pgs, mps[m]...)
					}
				}
			} else {
				return nil, fmt.Errorf("plugin %s not found", plugin)
			}
		}
	}

	return pgs, nil
}
