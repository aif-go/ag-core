package dao

import (
 db "github.com/aif-go/ag-core/contribute/agdb/gormdb"
 "github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

// TmTeacherNamingSqlMap 命名SQL映射
var TmTeacherNamingSqlMap = map[string]string{}

// excludeTmTeacherZeroColNames 插入忽略空值时标记哪些字段需要排除在外
var excludeTmTeacherZeroColNames = map[string]int{"JpaVersion": 0, "CreateTime": 0, "LastUpdateTime": 0}



var FindByPhoneNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "FindByPhone",
	ReqType:  (*model.TmTeacherFindByPhoneArg)(nil),
	RespType: (*model.TmTeacherFindByPhonePageRes)(nil),
}


var FindByNameNadAddressNamingInfo = &db.NameingSqlArgInfo{
	SqlName:  "FindByNameNadAddress",
	ReqType:  (*model.TmTeacherFindByNameNadAddressArg)(nil),
	RespType: ([]*model.TmTeacherFindByNameNadAddressRes)(nil),
}
