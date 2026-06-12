package config

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"github.com/aif-go/ag-core/ag/ag_conf/reader"
	"github.com/aif-go/ag-core/ag/ag_ext"
	"fmt"
	"log/slog"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

const (
	SourceKeyNacosPrefix string = "[NACOS]"
)

func EnableNacosRemoteConfig(env ag_conf.IConfigurableEnvironment, iClient config_client.IConfigClient, p *NacosConfigProperties) error {
	if p == nil || !p.Enable {
		slog.Info("nacos remote config is disable")
		return nil
	}

	dateids := p.DataIDs
	dataidCount := len(dateids)
	if dataidCount < 1 {
		slog.Info("nacos remote config dataids is empty")
		return nil
	}

	for i := dataidCount - 1; i >= 0; i-- {
		dataidinfo := dateids[i]

		// for _, dataidinfo := range dateids {
		if dataidinfo.DataID == "" {
			return fmt.Errorf("nacos config must config dataid value")
		}
		if dataidinfo.Group == "" {
			return fmt.Errorf("nacos config must config group value")
		}

		var content string
		content, err := iClient.GetConfig(vo.ConfigParam{
			DataId: dataidinfo.DataID,
			Group:  dataidinfo.Group,
			// Content: context,
			// Type:    vo.YAML,
		})
		if err != nil {
			return fmt.Errorf("dataId:%s Group:%s get config error: %w", dataidinfo.DataID, dataidinfo.Group, err)
		}
		if content == "" {
			slog.Info("nacos config is empty", "dataId:", dataidinfo.DataID, "Group:", dataidinfo.Group)
			continue
		}

		nacosSource, err := BuildNacosConfigPropertySource(&dataidinfo, content)
		if err != nil {
			return err
		}

		pss := env.GetPropertySources()
		if pss.ContainsSource(nacosSource) {
			slog.Info(fmt.Sprintf("nacos config already exists, dataId:%s Group:%s", dataidinfo.DataID, dataidinfo.Group))
			err = pss.ReplaceSource(nacosSource)
			if err != nil {
				return fmt.Errorf("dataId:%s Group:%s replace error: %w", dataidinfo.DataID, dataidinfo.Group, err)
			}
		} else {
			slog.Info(fmt.Sprintf("nacos config, dataId:%s Group:%s", dataidinfo.DataID, dataidinfo.Group))
			err := pss.AddAfter(ag_conf.SourceKeySystemEnvironment, nacosSource) // nacos配置优先级在环境变量之后
			if err != nil {
				return err
			}
		}

		// 解密nacos配置
		err = ag_conf.CreateOrUpdateDecryptForPropertySource(env, nacosSource)
		if err != nil {
			return err
		}

		// 注册nacos配置监听
		registerNacosWatcherIfNeed(dataidinfo, iClient)
	}
	return nil
}

// NacosPropertySource nacos配置实体
type NacosPropertySource struct {
	ag_conf.MapPropertySource
}

// NewNacosPropertySource naocs远程配置相关内容 当前是main方法主动放入env
func NewNacosPropertySource(name string, source map[string]any) *NacosPropertySource {
	return &NacosPropertySource{
		MapPropertySource: ag_conf.MapPropertySource{
			NamedPropertySource: ag_conf.NamedPropertySource{
				Name: name,
			},
			Source: source,
		},
	}

}

func getSourceKey(dataidinfo *DataIDInfo) string {
	return fmt.Sprintf("%s-%s_%s", SourceKeyNacosPrefix, dataidinfo.Group, dataidinfo.DataID)
}

// 若需要添加watcher
func registerNacosWatcherIfNeed(info DataIDInfo, iClient config_client.IConfigClient) {
	if !info.AutoRefresh {
		slog.Info("nacos config not auto refresh", "dataId:", info.DataID, "Group:", info.Group)
		return
	}

	slog.Info("nacos config auto refresh", "dataId:", info.DataID, "Group:", info.Group)

	// 复制一份,防止被修改
	winfo := &DataIDInfo{
		DataID:      info.DataID,
		Group:       info.Group,
		Type:        info.Type,
		AutoRefresh: info.AutoRefresh,
	}
	watcher := NewNacosConfigWatcher(winfo, iClient)

	ag_conf.RegisterWatcher(watcher)
}

func BuildNacosConfigPropertySource(dataidinfo *DataIDInfo, content string) (*NacosPropertySource, error) {
	key := getSourceKey(dataidinfo)
	cty := dataidinfo.Type
	reader, ok := reader.Readers[cty]
	if !ok {
		return nil, fmt.Errorf("fileType:%s not be supported", cty)
	}

	contextMap, err := reader([]byte(content))
	if err != nil {
		return nil, fmt.Errorf("dataId:%s Group:%s read error: %w", dataidinfo.DataID, dataidinfo.Group, err)
	}

	flatmapcontext, err := ag_ext.GetFlattenedMap(contextMap)
	if err != nil {
		return nil, err
	}

	nacosSource := NewNacosPropertySource(key, flatmapcontext)
	return nacosSource, nil
}
