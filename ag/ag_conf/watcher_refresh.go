package ag_conf

import (
	"encoding/json"
	"fmt"
	"log/slog"
)

func (wm *WatcherManager) refreshPropertySources(propertySources []IPropertySource) {
	// TODO 刷新和watcher close的冲突如何兼容

	// log
	psnames := make([]string, 0)
	for _, ps := range propertySources {
		psnames = append(psnames, ps.GetName())
	}
	slog.Info("refreshPropertySources", slog.Any("names", psnames))

	wm.doRefreshPropertySource(propertySources)
}

func (wm *WatcherManager) doRefreshPropertySource(pss []IPropertySource) {
	env := wm.env
	// 1. 获取所有有变化的key
	changeskv := make(map[string]interface{})
	for _, ps := range pss {
		psbefor := env.GetPropertySources().Get(ps.GetName())
		if psbefor == nil {
			slog.Warn(fmt.Sprintf("refreshPropertySources not found propertySource name=%s, will be ignore", ps.GetName()))
			return
		}
		befor := psbefor.GetSource()
		after := ps.GetSource()
		changs := changes(befor, after)
		for key, val := range changs {
			changeskv[key] = val
		}
		env.GetPropertySources().ReplaceSource(ps) // 更新配置源
	}
	cjson, _ := json.MarshalIndent(changeskv, "", " ")
	slog.Info(fmt.Sprintf("doRefreshPropertySource changes:\n%s", cjson))

	// 2. 获取所有变化的key的listener
	wm.refreshMapLock.RLock()
	defer wm.refreshMapLock.RUnlock()
	invokListeners := make(map[string][]func(k, v string))

	for key, listeners := range wm.refreshMap {
		// TODO key的比对要考虑slice类型的key识别
		if _, ok := changeskv[key]; ok {
			invokListeners[key] = listeners
		}
	}

	// 3. 执行所有变化的key的listener
	for key, listeners := range invokListeners {
		for _, listener := range listeners {
			// 判断key是否为slice格式
			listener(key, "")
		}
	}

}

func changes(befor, after map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// 检查before中有而after中没有的key，或者值不同的key
	for key, val := range befor {
		if newVal, exists := after[key]; !exists {
			result[key] = nil // 移除的key
		} else if val != newVal {
			result[key] = newVal // 修改的key
		}
	}

	// 检查after中有而before中没有的key
	for key, val := range after {
		if _, exists := befor[key]; !exists {
			result[key] = val // 新增的key
		}
	}

	return result
}
