package dao
// TmEtcRelationshipNamingSqlMap 命名SQL映射
var TmEtcRelationshipNamingSqlMap = map[string]string{}

// excludeTmEtcRelationshipZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmEtcRelationshipZeroColNames = map[string]int{"CreatedTime": 0, "LastModifiedTime": 0}

