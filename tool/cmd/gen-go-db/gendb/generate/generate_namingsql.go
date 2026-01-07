package generate

import (
	"regexp"
	"strings"
)

// 预编译正则表达式
var partten = regexp.MustCompile(` @[a-z|A-Z]+ *`)

func NamingSql(namingsql string)[]string{
	result:=partten.FindAllString(namingsql,-1)
	// result:=partten.FindStringSubmatch(namingsql)
	var goNameArr []string
	existMap:=make(map[string]int,5)
	if len(result)>0{
		goNameArr =[]string{}
		for _,paramter:=range result{
			paramter=strings.ReplaceAll(strings.TrimSpace(paramter),"@","")
			if _,exists:=existMap[paramter];!exists{
				goNameArr=append(goNameArr, paramter)
			}
			// log.Println("正则表达式结果:",paramter)
		}
	}
	return goNameArr
}