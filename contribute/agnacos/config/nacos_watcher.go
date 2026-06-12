package config

import (
	"github.com/aif-go/ag-core/ag/ag_conf"
	"context"
	"log/slog"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

//	type WatchInfo struct {
//		DataID      string
//		Group       string
//		Type        string
//		AutoRefresh bool
//	}
type NacosConfigWatcher struct {
	winfo   *DataIDInfo
	iClient config_client.IConfigClient
}

func NewNacosConfigWatcher(info *DataIDInfo, iClient config_client.IConfigClient) *NacosConfigWatcher {
	return &NacosConfigWatcher{
		winfo:   info,
		iClient: iClient,
	}
}

func (w *NacosConfigWatcher) Start(ctx context.Context, doChange ag_conf.ChangePropertySources) {
	winfo := w.winfo
	w.iClient.ListenConfig(vo.ConfigParam{
		DataId: winfo.DataID,
		Group:  winfo.Group,
		Type:   vo.ConfigType(winfo.Type), // 不指定类型能拿到吗
		OnChange: func(namespace string, group string, dataId string, data string) {
			slog.Info("nacos conf change", "ns:", namespace, "group:", group, "dataId:", dataId)
			ps, err := BuildNacosConfigPropertySource(winfo, data)
			if err != nil {
				slog.Error("nacos conf refresh", "dataId:", dataId, " errormsg:", err.Error())
				return
			}

			// refreshChan <- []ag_conf.IPropertySource{ps} // 将新PropertySource送给参数刷新处理
			doChange([]ag_conf.IPropertySource{ps})
		},
	})

}

func (w *NacosConfigWatcher) Stop() {
	w.iClient.CancelListenConfig(vo.ConfigParam{
		DataId: w.winfo.DataID,
		Group:  w.winfo.Group,
	})
}
