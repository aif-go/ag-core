package ag_ext

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// GetFlattenedMap 构建扁平化的map
// author: hzw
// date: 2025-05-01
// 递归深度可通过环境变量FLATTENED_MAP_MAX_DEPTH指定，默认100
func GetFlattenedMap(source interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	depth := 100

	// 从环境变量获取
	if envDepth := os.Getenv("FLATTENED_MAP_MAX_DEPTH"); envDepth != "" {
		if d, err := strconv.Atoi(envDepth); err == nil && d > 0 {
			depth = d
		}
	}

	err := buildFlattenedMap(result, source, "", depth)
	return result, err
}

// buildFlattenedMap 递归构建扁平化的map
// author: hzw
// date: 2025-05-01
func buildFlattenedMap(result map[string]interface{}, source interface{}, path string, maxDepth int) error {
	if result == nil || source == nil {
		return fmt.Errorf("nil map provided")
	}

	if strings.Count(path, ".") > maxDepth {
		return fmt.Errorf("recursion depth exceeded maximum limit of %d", maxDepth)
	}

	var sourceMap map[string]interface{}
	switch s := source.(type) {
	case map[string]interface{}:
		sourceMap = s
	case map[any]any:
		sourceMap = make(map[string]interface{})
		for k, v := range s {
			if strKey, ok := k.(string); ok {
				sourceMap[strKey] = v
			} else {
				sourceMap[fmt.Sprintf("%v", k)] = v
			}
		}
	default:
		return fmt.Errorf("unsupported source type: %T", source)
	}

	for key, value := range sourceMap {
		if key == "" {
			return fmt.Errorf("empty key found in source map")
		}
		newKey := key
		if path != "" {
			if strings.HasPrefix(key, "[") {
				newKey = path + key
			} else {
				newKey = path + "." + key
			}
		}

		switch v := value.(type) {
		case string:
			result[newKey] = v
		case map[string]interface{}, map[any]any:
			if err := buildFlattenedMap(result, v, newKey, maxDepth); err != nil {
				return fmt.Errorf("failed to flatten map for key %s: %w", newKey, err)
			}
		case []interface{}:
			for i, item := range v {
				indexKey := fmt.Sprintf("[%d]", i)
				if err := buildFlattenedMap(result, map[string]interface{}{indexKey: item}, newKey, maxDepth); err != nil {
					return fmt.Errorf("failed to flatten array item %d for key %s: %w", i, newKey, err)
				}
			}
		default:
			if v != nil {
				result[newKey] = fmt.Sprintf("%v", v)
			} else {
				result[newKey] = ""
			}
		}
	}
	return nil
}

// MergeMap 合并map
// func MergeMap(src map[string]interface{}, target map[string]interface{}) {
// 	if src != nil {
// 		for key, value := range src {
// 			// target中存在就覆盖
// 			target[key] = value
// 		}
// 	}
// }
