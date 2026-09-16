package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

const DB2_TmTeacher_FindByNameNadAddress = "SELECT name,address,phone,class_id FROM tm_teacher WHERE (name = @Name AND address = @Address)"

const DB2_TmTeacher_FindByPhone = "SELECT * FROM (SELECT *, ROW_NUMBER() OVER(ORDER BY id) AS RN  FROM tm_teacher WHERE (phone = @Phone)) AS T WHERE RN BETWEEN @Start AND @End"

const DB2_TmTeacher_FindByPhone_Count = "SELECT COUNT(*) FROM tm_teacher WHERE (phone = @Phone)"

func InitTmTeacherDB2() {
	TmTeacherNamingSqlMap["DB2_TmTeacher_FindByNameNadAddress"] = DB2_TmTeacher_FindByNameNadAddress
	TmTeacherNamingSqlMap["DB2_TmTeacher_FindByPhone"] = DB2_TmTeacher_FindByPhone
	TmTeacherNamingSqlMap["DB2_TmTeacher_FindByPhone_Count"] = DB2_TmTeacher_FindByPhone_Count
}
