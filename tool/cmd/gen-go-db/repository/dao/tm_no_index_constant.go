package dao
// TmNoIndexNamingSqlMap 命名SQL映射
var TmNoIndexNamingSqlMap = map[string]string{}

// excludeTmNoIndexZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmNoIndexZeroColNames = map[string]int{"JpaVersion": 0, "CreateTime": 0, "LastUpdateTime": 0}

