package ag_conf

import "unicode"

/* 内部使用 */

// hasPrefixIgnoreCase 检查字符串s是否以prefix开头，忽略大小写
// 进一步优化：对ASCII字符使用更快的比较方式
func hasPrefixIgnoreCase(s, prefix string) bool {
	if len(prefix) > len(s) {
		return false
	}

	for i := 0; i < len(prefix); i++ {
		sc := s[i]
		pc := prefix[i]

		// 针对ASCII字符的快速比较
		if sc >= 'A' && sc <= 'Z' {
			sc += 32 // 转为小写
		}
		if pc >= 'A' && pc <= 'Z' {
			pc += 32 // 转为小写
		}

		// 非ASCII字符使用unicode.ToLower
		if sc > 127 || pc > 127 {
			if unicode.ToLower(rune(sc)) != unicode.ToLower(rune(pc)) {
				return false
			}
		} else if sc != pc {
			return false
		}
	}
	return true
}

// TrimPrefixIgnoreCase 高性能版本：移除字符串s开头的prefix（忽略大小写）
func trimPrefixIgnoreCase(s, prefix string) string {
	// 处理空前缀情况
	if prefix == "" {
		return s
	}

	prefixLen, sLen := len(prefix), len(s)
	// 前缀更长时直接返回原字符串
	if prefixLen > sLen {
		return s
	}

	// 逐个字符比较（优先处理ASCII字符提升性能）
	for i := 0; i < prefixLen; i++ {
		sc, pc := s[i], prefix[i]

		// 先检查是否为相同字符（不区分大小写的快速路径）
		if sc == pc {
			continue
		}

		// 处理ASCII字母的大小写转换（比unicode包更快）
		if sc >= 'A' && sc <= 'Z' {
			sc += 32 // 转小写
		}
		if pc >= 'A' && pc <= 'Z' {
			pc += 32 // 转小写
		}

		// 处理非ASCII字符
		if sc > 127 || pc > 127 {
			scLower := unicode.ToLower(rune(sc))
			pcLower := unicode.ToLower(rune(pc))
			if scLower != pcLower {
				return s
			}
			// 确保后续字节也匹配
			if scLower == pcLower && len(string(scLower)) > 1 {
				i += len(string(scLower)) - 1
			}
		} else if sc != pc {
			return s
		}
	}

	// 所有字符匹配，返回移除前缀后的子串
	return s[prefixLen:]
}
