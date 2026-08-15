package dao
// TmNoNamingSqlMap 命名SQL映射
var TmNoNamingSqlMap = map[string]string{}

// excludeTmNoZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmNoZeroColNames = map[string]int{"JpaVersion": 0, "CreateTime": 0, "LastUpdateTime": 0}

