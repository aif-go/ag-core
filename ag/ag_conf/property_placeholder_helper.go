package ag_conf

import (
	"fmt"
	"strings"
	"sync"
)

var pphBuilderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

type PlaceholderResolver func(placeholder string) string

type PropertyPlaceholderHelper struct {
	PlaceholderPrefix              string
	PlaceholderSuffix              string
	ValueSeparator                 string
	IgnoreUnresolvablePlaceholders bool //
}

func NewPropertyPlaceholderHelper(placeholderPrefix string, placeholderSuffix string, valueSeparator string, ignoreUnresolvablePlaceholders bool) *PropertyPlaceholderHelper {
	return &PropertyPlaceholderHelper{
		PlaceholderPrefix:              placeholderPrefix,
		PlaceholderSuffix:              placeholderSuffix,
		ValueSeparator:                 valueSeparator,
		IgnoreUnresolvablePlaceholders: ignoreUnresolvablePlaceholders,
	}
}

func (ph *PropertyPlaceholderHelper) ReplacePlaceholders(value string, placeholderResolver PlaceholderResolver) (string, error) {
	if value == "" {
		return "", fmt.Errorf("value is empty")
	}
	return ph.parseStringValue(value, placeholderResolver)
}

func (ph *PropertyPlaceholderHelper) parseStringValue(value string, placeholderResolver PlaceholderResolver) (string, error) {
	return ph.parseStringValueWithVisited(value, placeholderResolver, make(map[string]bool))
}

/**
 * parseStringValueWithVisited
 * 递归解析字符串中的占位符，并替换为对应的值。
 * 支持嵌套占位符和默认值，防止循环引用。
 *
 * @param value 待解析的字符串，可能包含占位符
 * @param placeholderResolver 用于解析占位符对应值的函数
 * @param visitedPlaceholders 记录当前解析路径中已访问的占位符，防止循环引用
 * @return 解析后的字符串及可能的错误
 */
func (ph *PropertyPlaceholderHelper) parseStringValueWithVisited(
	value string,
	placeholderResolver PlaceholderResolver,
	visitedPlaceholders map[string]bool,
) (string, error) {
	prefixLen := len(ph.PlaceholderPrefix)
	suffixLen := len(ph.PlaceholderSuffix)
	separatorLen := len(ph.ValueSeparator)

	builder := pphBuilderPool.Get().(*strings.Builder)

	defer func() {
		builder.Reset()
		pphBuilderPool.Put(builder)
	}()

	// builder := &strings.Builder{}
	builder.Reset()
	builder.Grow(len(value) * 2) // 预分配足够空间
	builder.WriteString(value)

	startIndex := strings.Index(builder.String(), ph.PlaceholderPrefix)
	for startIndex != -1 {
		endIndex := ph.findPlaceholderEndIndex(builder.String(), startIndex)
		if endIndex == -1 {
			break // 未找到对应的后缀，停止解析
		}

		placeholder := builder.String()[startIndex+prefixLen : endIndex]
		originalPlaceholder := placeholder

		// 检测循环引用
		if visitedPlaceholders[originalPlaceholder] {
			return "", fmt.Errorf("circular placeholder reference '%s' in property definitions", originalPlaceholder)
		}
		visitedPlaceholders[originalPlaceholder] = true

		// 递归解析占位符中的嵌套占位符，最终获取到占位符KEY
		resolvedPlaceholder, err := ph.parseStringValueWithVisited(placeholder, placeholderResolver, visitedPlaceholders)
		if err != nil {
			return "", err
		}

		// 通过解析器获取占位符对应的值
		propVal := placeholderResolver(resolvedPlaceholder)

		if propVal == "" && ph.ValueSeparator != "" {
			// 支持默认值分隔符，如 ${key:defaultValue}
			if sepIdx := strings.Index(resolvedPlaceholder, ph.ValueSeparator); sepIdx != -1 {
				// 分割出实际key
				actualPlaceholder := resolvedPlaceholder[:sepIdx]
				// 默认值
				defaultValue := resolvedPlaceholder[sepIdx+separatorLen:]
				// 获取实际key对应值
				propVal = placeholderResolver(actualPlaceholder)
				// 若值为空则使用默认值
				if propVal == "" {
					propVal = defaultValue
				}
			}
		}

		if propVal != "" {
			// 已获取到值，递归解析替换值中的占位符
			resolvedValue, err := ph.parseStringValueWithVisited(propVal, placeholderResolver, visitedPlaceholders)
			if err != nil {
				return "", err
			}

			// 直接操作Builder进行替换
			newStr := builder.String()[:startIndex] + resolvedValue + builder.String()[endIndex+suffixLen:]
			builder.Reset()
			builder.WriteString(newStr)

			// 更新搜索位置，搜索起始占位符
			searchPos := startIndex + len(resolvedValue)
			if nextIdx := strings.Index(builder.String()[searchPos:], ph.PlaceholderPrefix); nextIdx != -1 {
				startIndex = searchPos + nextIdx
			} else {
				startIndex = -1
			}
		} else if ph.IgnoreUnresolvablePlaceholders {
			// 忽略无法解析的占位符，继续查找下一个占位符
			searchPos := endIndex + suffixLen
			if nextIdx := strings.Index(builder.String()[searchPos:], ph.PlaceholderPrefix); nextIdx != -1 {
				startIndex = searchPos + nextIdx
			} else {
				startIndex = -1
			}
		} else {
			// 不允许忽略，返回错误
			return "", fmt.Errorf("could not resolve placeholder '%s' in value '%s'", resolvedPlaceholder, value)
		}

		delete(visitedPlaceholders, originalPlaceholder)
	}

	// 没有占位符则直接返回原值
	return builder.String(), nil
}

// findPlaceholderEndIndex 查找占位符的结束索引
func (ph *PropertyPlaceholderHelper) findPlaceholderEndIndex(buf string, startIndex int) int {
	index := startIndex + len(ph.PlaceholderPrefix)
	withinNestedPlaceholder := 0

	for index < len(buf) {
		if strings.HasPrefix(buf[index:], ph.PlaceholderSuffix) {
			if withinNestedPlaceholder > 0 {
				withinNestedPlaceholder--
				index += len(ph.PlaceholderSuffix)
			} else {
				return index
			}
		} else if strings.HasPrefix(buf[index:], ph.PlaceholderPrefix) {
			withinNestedPlaceholder++
			index += len(ph.PlaceholderPrefix)
		} else {
			index++
		}
	}
	return -1
}
