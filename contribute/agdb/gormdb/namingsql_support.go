package gormdb

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// 预编译正则表达式，避免每次调用时重新编译
var (
	// 匹配 in ( @XXX ) 和 not in ( @XXX ) 格式
	inRegex = regexp.MustCompile(`(?i)\s+(not\s+)?in\s*\(\s*@([A-Za-z0-9_]+)\s*\)`)
	// 匹配其他 @XXX 格式的命名参数
	namedParamRegex = regexp.MustCompile(`@([A-Za-z0-9_]+)`)
)

// ReplaceNamedParamsWithIn 该方法用于替换 SQL 中的命名参数，特别处理 in ( @XXX ) 为 in ?
// 优化点：
// 1. 预编译正则表达式，提高性能
// 2. 支持 not in ( @XXX ) 格式
// 3. 优化正则表达式，减少回溯
// 4. 统一处理参数名，避免重复记录
func ReplaceNamedParamsWithIn(sql string) (replacedSQL string, paramNames []string) {
	// 1. 处理 in ( @XXX ) 和 not in ( @XXX ) 格式
	var inParamNames []string
	replacedSQL = inRegex.ReplaceAllStringFunc(sql, func(match string) string {
		// 提取参数名（匹配结果的第三个分组）
		matches := inRegex.FindStringSubmatch(match)
		if len(matches) > 2 {
			paramName := matches[2]
			inParamNames = append(inParamNames, paramName)
			// 保留 not 关键字（如果存在），替换为 in ?
			if len(matches[1]) > 0 {
				return " not in ?"
			}
			return " in ?"
		}
		return match
	})

	// 2. 处理其他命名参数
	var otherParamNames []string
	replacedSQL = namedParamRegex.ReplaceAllStringFunc(replacedSQL, func(match string) string {
		// 提取参数名（匹配结果的第二个分组）
		paramName := match[1:] // 去掉 @ 符号
		otherParamNames = append(otherParamNames, paramName)
		return "?"
	})

	// 3. 合并参数名列表
	paramNames = append(inParamNames, otherParamNames...)

	return
}

// // ReplaceNamedParamsWithIn 该方法对于generate也是使用的，自动替换 SQL 中的命名参数，特别处理 in ( @XXX ) 为 in ?
// func ReplaceNamedParamsWithIn(sql string) (replacedSQL string, paramNames []string) {
// 	// 1. 先处理 in ( @XXX ) 格式：匹配 "in ( @参数名 )"
// 	inRegex := regexp.MustCompile(`in\s*\(\s*@([A-Za-z0-9_]+)\s*\)`)
// 	// 记录 in 条件中的参数名
// 	inMatches := inRegex.FindAllStringSubmatch(sql, -1)
// 	for _, m := range inMatches {
// 	  paramNames = append(paramNames, m[1]) // 记录 @XXX 中的参数名（如 "AppIdSlice"）
// 	}
// 	// 将 "in ( @XXX )" 替换为 "in ?"
// 	sql = inRegex.ReplaceAllString(sql, "in ?")

// 	// 2. 处理其他命名参数（@XXX）
// 	otherRegex := regexp.MustCompile(`@([A-Za-z0-9_]+)`)
// 	otherMatches := otherRegex.FindAllStringSubmatch(sql, -1)
// 	for _, m := range otherMatches {
// 	  paramNames = append(paramNames, m[1])
// 	}
// 	// 将其他 @XXX 替换为 ?
// 	replacedSQL = otherRegex.ReplaceAllString(sql, "?")

// 	return
//   }

// GetParamsByNames 该方法对于generate也是使用的，根据参数名列表，从结构体中提取对应字段的值（按顺序）
func GetParamsByNames(arg interface{}, paramNames []string) ([]interface{}, error) {
	val := reflect.Indirect(reflect.ValueOf(arg))
	if val.Kind() != reflect.Struct {
		return nil, errors.New("arg must be a struct or pointer to struct")
	}

	params := make([]interface{}, 0, len(paramNames))
	for _, name := range paramNames {
		// 查找结构体中与参数名匹配的字段（大小写敏感，需与结构体字段名一致）
		field := val.FieldByName(name)
		if !field.IsValid() {
			return nil, fmt.Errorf("struct has no field: %s", name)
		}
		params = append(params, field.Interface())
	}
	return params, nil
}

// CalcPageStartRecord 计算分页查询的开始记录
// 如果开始记录大于总数，返回0, 不让查询
// 如果结束记录大于总数，返回总数
func CalcPageStartRecord(pageNum int64, pageSize int64, totalCount int64, dbType string) (int64, int64, int64) {
	// 1. 校验非法参数（提前拦截无效请求）
	if pageNum <= 0 {
		return 0, 0, 0 // 页码必须 ≥1
	}
	if pageSize <= 0 {
		return 0, 0, 0 // 页大小必须 ≥1
	}
	if totalCount == 0 {
		return 0, 0, 0 // 无数据，无需分页
	}

	// between and 左右区间都是闭合的 []，因此开始索引需要在 pageSize 基础上 +1
	var startRecord int64
	var endRecord int64

	switch dbType {
	case "mysql", "MYSQL":
		startRecord = (pageNum - 1) * pageSize
		endRecord = pageSize
	default:
		startRecord = (pageNum-1)*pageSize + 1
		endRecord = startRecord + pageSize - 1
	}
	if startRecord > totalCount {
		return 0, 0, 0
	}
	if endRecord > totalCount {
		endRecord = totalCount
	}
	// 计算总页数，向上取整
	totalPage := totalCount / pageSize
	if totalCount%pageSize != 0 {
		totalPage++
	}
	return startRecord, endRecord, totalPage
}

// collectZeroValWithOmitEmpty 收集：有 json omitempty 标记 + 值为零值 的字段名
func CollectZeroValWithOmitEmpty(obj interface{}) []string {
	var result []string
	// 1. 解析入参：支持结构体或结构体指针
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem() // 指针解引用，获取底层结构体
	}
	if val.Kind() != reflect.Struct {
		return result // 非结构体/指针直接返回空
	}
	typ := val.Type() // 获取结构体类型（用于解析 tag）
	// 2. 遍历字段：判断 tag 和值
	for i := 0; i < typ.NumField(); i++ {
		fieldTyp := typ.Field(i) // 字段类型（含 tag）
		fieldVal := val.Field(i) // 字段实际值
		// 2.1 判断是否有 gorm omitempty 标记
		// gormTag := fieldTyp.Tag.Get("gorm")
		// hasOmitEmpty := false
		// for _, opt := range strings.Split(gormTag, ";") {
		//     if opt == "omitempty" {
		//         hasOmitEmpty = true
		//         break
		//     }
		// }
		// if !hasOmitEmpty {
		//     continue // 没有 omitempty 标记，跳过
		// }
		// 2.2 判断字段值是否为零值
		if fieldVal.IsZero() {
			result = append(result, fieldTyp.Name) // 满足条件，收集字段名
		}
	}

	return result
}
