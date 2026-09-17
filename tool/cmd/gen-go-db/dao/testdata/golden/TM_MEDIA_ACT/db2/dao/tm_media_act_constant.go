package dao

import (
	db "github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

// TmMediaActNamingSqlMap 命名SQL映射
var TmMediaActNamingSqlMap = map[string]string{}

// excludeTmMediaActZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmMediaActZeroColNames = map[string]int{"CreatedTime": 0, "LastModifiedTime": 0}

var NoPageQueryNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "NoPageQuery",
	ReqType:  (*model.TmMediaActNoPageQueryArg)(nil),
	RespType: ([]*model.TmMediaActNoPageQueryRes)(nil),
}

var XxxxxNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "Xxxxx",
	ReqType:  (*model.TmMediaActXxxxxArg)(nil),
	RespType: (*model.TmMediaActXxxxxPageRes)(nil),
}
