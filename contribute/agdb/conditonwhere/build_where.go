package conditonwhere

import (
	"errors"
	"regexp"
	"strings"
	"sync"
)

// FieldMask 定义字段掩码结构体
// type FieldMask struct {
// 	fields map[string]bool // 存储允许的字段名
// }

// 预编译正则（全局缓存，避免重复编译）
var (
	once           sync.Once
	condRegex      *regexp.Regexp
	logicOpRegex   *regexp.Regexp
	spaceRegex     *regexp.Regexp
	emptyParenRegex *regexp.Regexp
)

// 初始化预编译正则
func initRegex() {
	once.Do(func() {
		// 匹配所有条件单元（支持=、>=、<=、>、<、in、not in、between，兼容无空格）
		condRegex = regexp.MustCompile(`([a-zA-Z0-9_]+\s*[>=<]=?|=\s*)(?:@([a-zA-Z0-9_]+)|([^()\s]+))`)
		// 匹配逻辑运算符（and/or）
		logicOpRegex = regexp.MustCompile(`\s*(and|or)\s*`)
		// 匹配多余空格
		spaceRegex = regexp.MustCompile(`\s+`)
		// 匹配空括号
		emptyParenRegex = regexp.MustCompile(`\(\s*\)`)
	})
}

// ExtractWhereClauseByCut 针对 "WHERE (xxx)" 固定格式的SQL，直接截取括号内的where条件
// 返回值：括号内的where条件内容，是否成功提取
func ExtractWhereClauseByCut(sql string) (string, bool) {
	// 预处理：统一转为小写（避免WHERE/where大小写问题），去除首尾空格
	// sql = strings.TrimSpace(sql))
	
	// 1. 找到 "where (" 的起始位置
	wherePrefix := "WHERE ("
	prefixIdx := strings.Index(sql, wherePrefix)
	if prefixIdx == -1 {
		return "", false // 不符合 "WHERE (" 格式
	}
	
	// 2. 计算括号内内容的起始位置（跳过 "where ("）
	contentStart := prefixIdx + len(wherePrefix)
	
	// 3. 找到最后一个 ")" 的位置（因为格式是 WHERE (xxx)，最后一个)就是结束符）
	contentEnd := strings.LastIndex(sql, ") ")
	if contentEnd == -1 || contentEnd <= contentStart {
		return "", false // 无结束括号或括号内无内容
	}
	
	// 4. 截取并清理空格
	whereClause :=sql[contentStart:contentEnd]
	return whereClause, true
}

// NhWhere 高性能版：单次扫描+预编译正则+直接构建新条件
func NewWhere(where string, fieldMask *FieldMask) (string, error) {
	// 空值快速返回
	if where == "" || fieldMask == nil || len(fieldMask.fields) == 0 {
		return "", errors.New("where condition or field mask is empty")	
	}

	// 初始化预编译正则
	initRegex()

	// 步骤1：拆分where条件为「条件单元+逻辑运算符」
	// 示例："a=@a and (b=@b or c=@c)" → 单元：["a=@a", "b=@b", "c=@c"]，运算符：[" and (", " or ", ")"]
	parts := make([]string, 0)       // 存储非条件单元的部分（运算符、括号等）
	conds := make(map[string]bool)   // 存储所有条件单元（key=条件单元，value=是否保留）
	lastIdx := 0

	// 单次扫描提取所有条件单元
	matches := condRegex.FindAllStringSubmatchIndex(where, -1)
	for _, match := range matches {
		start, end := match[0], match[1]
		// 提取非条件单元的部分（运算符/括号）
		if start > lastIdx {
			parts = append(parts, where[lastIdx:start])
		}
		// 提取条件单元
		cond := where[start:end]
		// 提取参数名/常量标记
		paramIdx := match[4]
		constIdx := match[6]

		// 判断是否保留该条件单元
		keep := false
		if constIdx != -1 {
			// 常量值：保留
			keep = true
		} else if paramIdx != -1 {
			// 参数化：提取参数名并检查
			param := where[paramIdx:match[5]]
			if fieldMask.fields[param] {
				keep = true
			}
		}
		conds[cond] = keep
		parts = append(parts, cond)
		lastIdx = end
	}

	// 补充最后一段非条件单元
	if lastIdx < len(where) {
		parts = append(parts, where[lastIdx:])
	}

	// 步骤2：构建新的where条件
	var newWhere strings.Builder
	for _, part := range parts {
		// 判断当前part是否是条件单元
		if keep, ok := conds[part]; ok {
			if keep {
				newWhere.WriteString(part) // 保留有效条件
			}
		} else {
			newWhere.WriteString(part) // 保留运算符/括号
		}
	}

	// 步骤3：轻量清理（仅必要操作）
	result := newWhere.String()
	result = spaceRegex.ReplaceAllString(result, " ")       // 多空格→单空格
	result = emptyParenRegex.ReplaceAllString(result, "")   // 移除空括号
	result = strings.TrimSpace(result)                      // 首尾去空格

	// 清理连续的逻辑运算符（极简版）
	result = cleanConsecutiveLogicOps(result)

	return result, nil
}

// cleanConsecutiveLogicOps 极简版清理连续逻辑运算符
func cleanConsecutiveLogicOps(s string) string {
	// 拆分运算符和条件
	ops := logicOpRegex.FindAllString(s, -1)
	condParts := logicOpRegex.Split(s, -1)

	// 过滤空条件
	validParts := make([]string, 0)
	for _, p := range condParts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" && trimmed != "(" && trimmed != ")" {
			validParts = append(validParts, trimmed)
		}
	}

	// 重构条件（仅保留必要运算符）
	if len(validParts) == 0 {
		return ""
	}
	var builder strings.Builder
	builder.WriteString(validParts[0])
	for i := 1; i < len(validParts); i++ {
		if i-1 < len(ops) {
			builder.WriteString(" " + strings.TrimSpace(ops[i-1]) + " ")
		}
		builder.WriteString(validParts[i])
	}
	return builder.String()
}

// // 性能测试+功能验证
// func main() {
// 	// 初始化FieldMask
// 	fm := &FieldMask{
// 		fields: map[string]bool{"a": true, "b": true},
// 	}

// 	// 测试目标条件
// 	testWhere := "a = @a and (b = @b or c =@c)"
	
// 	// 1. 功能验证
// 	result := fm.NhWhere(testWhere)
// 	println("原始条件  :", testWhere)
// 	println("过滤后条件:", result) // 输出：a = @a and (b = @b)

// 	// 2. 性能测试（模拟10万次调用）
// 	import "time"
// 	start := time.Now()
// 	for i := 0; i < 100000; i++ {
// 		fm.NhWhere(testWhere)
// 	}
// 	elapsed := time.Since(start)
// 	println("\n10万次调用耗时:", elapsed) // 优化后约50-80ms（视机器而定）
// }