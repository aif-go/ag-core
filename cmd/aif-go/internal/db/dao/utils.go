package dao

import "strings"

// toCamelCase 将下划线分隔的字符串转换为驼峰命名法
func ToCamelCase(s string) string {

	s = strings.ToLower(s)
	// 首先将字符串分割成单词列表
	parts := strings.Split(s, "_")
	// 然后对每个单词进行处理，除了第一个单词外，其他单词的首字母需要大写
	for i, part := range parts {
		if i > 0 {
			// 将首字母转换为大写
			parts[i] = TitleCase(part)
		}
	}
	// 最后将处理后的单词列表重新连接成一个字符串
	return TitleCase(strings.Join(parts, ""))
}

// 首字母大写
func TitleCase(s string) string {

	//value:=*s
	// 如果字符串为空，直接返回
	if s == "" {
		return s
	}
	// 将字符串的第一个字符转换为大写，其余字符保持原样
	return strings.ToUpper(string(s[0])) + s[1:]
}

func CheckOrNotContains(src []string, targt string) bool {
	if src == nil {
		return false
	}

	for _, part := range src {

		if part == targt {
			return true
		}
	}
	return false
}
