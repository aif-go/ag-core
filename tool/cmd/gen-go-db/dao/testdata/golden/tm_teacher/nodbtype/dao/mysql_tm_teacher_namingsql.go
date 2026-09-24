package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

const MYSQL_TmTeacher_FindByNameNadAddress = "SELECT name,address,phone,class_id FROM tm_teacher WHERE (name = @Name AND address = @Address)"

const MYSQL_TmTeacher_FindByPhone = "SELECT * FROM tm_teacher WHERE (phone = @Phone) LIMIT @Start, @End"

const MYSQL_TmTeacher_FindByPhone_Count = "SELECT COUNT(*) FROM tm_teacher WHERE (phone = @Phone)"

func InitTmTeacherMYSQL() {
	TmTeacherNamingSqlMap["MYSQL_TmTeacher_FindByNameNadAddress"] = MYSQL_TmTeacher_FindByNameNadAddress
	TmTeacherNamingSqlMap["MYSQL_TmTeacher_FindByPhone"] = MYSQL_TmTeacher_FindByPhone
	TmTeacherNamingSqlMap["MYSQL_TmTeacher_FindByPhone_Count"] = MYSQL_TmTeacher_FindByPhone_Count
}
