package utils

import "strings"

// ParseCommaSeparatedList 解析逗号分隔的字符串列表
func ParseCommaSeparatedList(input string) []string {
	var result []string
	if input != "" {
		items := strings.Split(input, ",")
		// 去除每个项目的空格
		for i, item := range items {
			items[i] = strings.TrimSpace(item)
		}
		result = items
	}
	return result
}

// ContainsIgnoreCase 检查一个字符串是否在字符串列表中（不区分大小写）
func ContainsIgnoreCase(list []string, item string) bool {
	for _, str := range list {
		if strings.EqualFold(str, item) {
			return true
		}
	}
	return false
}
