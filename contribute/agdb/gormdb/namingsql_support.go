package gormdb

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
)

// ReplaceNamedParamsWithIn 该方法对于generate也是使用的，自动替换 SQL 中的命名参数，特别处理 in ( @XXX ) 为 in ?
func ReplaceNamedParamsWithIn(sql string) (replacedSQL string, paramNames []string) {
	// 1. 先处理 in ( @XXX ) 格式：匹配 "in ( @参数名 )"
	inRegex := regexp.MustCompile(`in\s*\(\s*@([A-Za-z0-9_]+)\s*\)`)
	// 记录 in 条件中的参数名
	inMatches := inRegex.FindAllStringSubmatch(sql, -1)
	for _, m := range inMatches {
	  paramNames = append(paramNames, m[1]) // 记录 @XXX 中的参数名（如 "AppIdSlice"）
	}
	// 将 "in ( @XXX )" 替换为 "in ?"
	sql = inRegex.ReplaceAllString(sql, "in ?")
  
	// 2. 处理其他命名参数（@XXX）
	otherRegex := regexp.MustCompile(`@([A-Za-z0-9_]+)`)
	otherMatches := otherRegex.FindAllStringSubmatch(sql, -1)
	for _, m := range otherMatches {
	  paramNames = append(paramNames, m[1])
	}
	// 将其他 @XXX 替换为 ?
	replacedSQL = otherRegex.ReplaceAllString(sql, "?")
  
	return
  }

  
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