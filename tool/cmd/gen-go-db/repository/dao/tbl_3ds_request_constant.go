package dao

import (
 db "github.com/aif-go/ag-core/contribute/agdb/gormdb"
 "github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

// Tbl3dsRequestNamingSqlMap 命名SQL映射
var Tbl3dsRequestNamingSqlMap = map[string]string{}

// excludeTbl3dsRequestZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTbl3dsRequestZeroColNames = map[string]int{"InsertTimestamp": 0}



var XxxxxNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "Xxxxx",
	ReqType:  (*model.Tbl3dsRequestXxxxxArg)(nil),
	RespType: ([]*model.Tbl3dsRequest)(nil),
}
