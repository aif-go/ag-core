package ag_conf

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
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
	// changeskv := make(map[string]interface{})
	changeskeys := make([]string, 0)
	for _, ps := range pss {
		psbefor := env.GetPropertySources().Get(ps.GetName())
		if psbefor == nil {
			slog.Warn(fmt.Sprintf("refreshPropertySources not found propertySource name=%s, will be ignore!!!", ps.GetName()))
			continue
		}
		befor := psbefor.GetSource()
		after := ps.GetSource()
		changs := changes(befor, after)
		for key, _ := range changs {
			// changeskv[key] = val
			changeskeys = append(changeskeys, key)
		}
		err := DecryptConfigSource(env, ps) // 密码解密Source更新
		if err != nil {
			slog.Warn(fmt.Sprintf("refreshPropertySources DecryptConfigSource err source:%s, will be ignore!!!, err:%v, ", ps.GetName(), err))
			continue
		}
		env.GetPropertySources().ReplaceSource(ps) // 更新配置源
	}
	cjson, _ := json.MarshalIndent(changeskeys, "", " ")
	slog.Info(fmt.Sprintf("doRefreshPropertySource changes:\n%s", cjson))

	// 2. 获取所有变化的key的listener
	wm.refreshMapLock.RLock()
	defer wm.refreshMapLock.RUnlock()
	invokListeners := make(map[string][]func(k, v string))

	for key, listeners := range wm.refreshMap {
		// TODO key的比对要考虑slice类型的key识别
		// if _, ok := changeskv[key]; ok {
		// 	invokListeners[key] = listeners
		// }
		for _, changekey := range changeskeys {
			// strings.EqualFold(changekey, key)
			// if strings.HasPrefix(strings.ToLower(changekey), strings.ToLower(key)) {
			if matchRefreshKey(changekey, key) {
				invokListeners[key] = listeners
			}
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

/*
// matchRefreshKey 匹配刷新key
// a.BB.c 匹配 a.BB.c a.bB
// a.BB.c 不匹配 a.B a.b
// a.b.c[1] 匹配 a.b.c[1] a.b.c
func matchRefreshKey(s, t string) bool {
	// 检查数组索引情况
	tHasIndex := strings.Contains(t, "[")
	sHasIndex := strings.Contains(s, "[")

	// 处理数组索引特殊情况
	if tHasIndex && !sHasIndex {
		// 如果t有数组索引而s没有，比较基础部分
		tBase := t[:strings.Index(t, "[")]
		return matchKeyWithoutIndex(s, tBase)
	} else if sHasIndex && !tHasIndex {
		// 如果s有数组索引而t没有，比较基础部分
		sBase := s[:strings.Index(s, "[")]
		return matchKeyWithoutIndex(sBase, t)
	} else if tHasIndex && sHasIndex {
		// 两者都有数组索引需要完全匹配
		return strings.EqualFold(s, t)
	}

	// 普通情况比较
	return matchKeyWithoutIndex(s, t)
}

// matchKeyWithoutIndex 处理不带数组索引的key比较
func matchKeyWithoutIndex(s, t string) bool {
	// 快速检查完全匹配
	if len(s) == len(t) && strings.EqualFold(s, t) {
		return true
	}

	// 检查前缀匹配
	if len(s) < len(t) {
		return false
	}

	// 比较前缀并检查是否是完整路径段
	return strings.EqualFold(s[:len(t)], t) &&
	       (len(s) == len(t) || s[len(t)] == '.')
}
*/

// matchRefreshKey 匹配刷新key
// 示例:
// - a.BB.c 匹配 a.BB.c a.bB
// - a.BB.c 不匹配 a.B a.b
// - a.b.c[1] 匹配 a.b.c[1] a.b.c
func matchRefreshKey(s, t string) bool {
	// 处理数组索引情况
	sBase, sHasIndex := splitIndex(s)
	tBase, tHasIndex := splitIndex(t)

	// 情况1: 两者都有索引，必须完全匹配（忽略大小写）
	if sHasIndex && tHasIndex {
		return strings.EqualFold(s, t)
	}

	// 情况2: 其中一个有索引，另一个没有，比较基础部分
	if sHasIndex || tHasIndex {
		return matchKeyWithoutIndex(sBase, tBase)
	}

	// 情况3: 两者都没有索引，直接比较
	return matchKeyWithoutIndex(s, t)
}

// splitIndex 分割路径和索引部分，返回基础路径和是否包含索引
func splitIndex(key string) (base string, hasIndex bool) {
	idx := strings.Index(key, "[")
	if idx == -1 {
		return key, false
	}
	return key[:idx], true
}

// matchKeyWithoutIndex 比较不带索引的路径部分
// 规则: s必须以t开头，且t必须是s的完整路径段（即t后为'.'或s结束）
func matchKeyWithoutIndex(s, t string) bool {
	if len(s) < len(t) {
		return false
	}

	// 检查前缀是否相等（忽略大小写）
	if !strings.EqualFold(s[:len(t)], t) {
		return false
	}

	// 检查t是否为s的完整路径段
	return len(s) == len(t) || s[len(t)] == '.'
}
