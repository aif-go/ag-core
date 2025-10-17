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

func RegPlugin(model, plugin string, pluginOpt string) {
	if _, ok := protoPluginsStor[model]; !ok {
		protoPluginsStor[model] = map[string][]string{}
	}

	if _, ok := protoPluginsStor[model][plugin]; !ok {
		protoPluginsStor[model][plugin] = []string{}
	}

	protoPluginsStor[model][plugin] = append(protoPluginsStor[model][plugin], pluginOpt)
}

// selectPlugins select plugins by models
func selectPlugins(plugins []string, models []string) ([]string, error) {

	pgs := []string{}

	if _, ok := protoPluginsStor[PluginBase]; ok {
		if _, ok := protoPluginsStor[PluginBase][ModelBase]; ok {
			pgs = append(pgs, protoPluginsStor[PluginBase][ModelBase]...)
		}
	}

	modelAll := lo.Contains(models, ModelAll)
	pluginAll := lo.Contains(plugins, PluginAll)

	// modelAll := slices.Contains(models, ModelAll)
	// pluginAll := slices.Contains(plugins, PluginAll)

	if pluginAll {

		// 加载所有插件
		for k, mps := range protoPluginsStor {
			if k == PluginBase {
				continue
			}

			for m, _ := range mps {
				if m == ModelBase {
					pgs = append(pgs, mps[m]...)
					continue
				}

				// if modelAll || slices.Contains(models, m) {
				if modelAll || lo.Contains(models, m) {
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
					if m == ModelBase {
						pgs = append(pgs, mps[m]...)
						continue
					}

					// if modelAll || slices.Contains(models, m) {
					if modelAll || lo.Contains(models, m) {
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
