package dao
// TmStudentNamingSqlMap 命名SQL映射
var TmStudentNamingSqlMap = map[string]string{}

// excludeTmStudentZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmStudentZeroColNames = map[string]int{"JpaVersion": 0, "CreateTime": 0, "LastUpdateTime": 0}

