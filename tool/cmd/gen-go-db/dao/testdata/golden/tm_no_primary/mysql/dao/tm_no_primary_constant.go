package dao
// TmNoPrimaryNamingSqlMap 命名SQL映射
var TmNoPrimaryNamingSqlMap = map[string]string{}

// excludeTmNoPrimaryZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmNoPrimaryZeroColNames = map[string]int{"JpaVersion": 0, "CreateTime": 0, "LastUpdateTime": 0}

