package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

const MYSQL_Tbl3dsRequest_Xxxxx = "SELECT * FROM tbl_3ds_request WHERE (RETRIEVAL_REFERENCE_NUMBER = @RRN)"

func InitTbl3dsRequestMYSQL() {
	Tbl3dsRequestNamingSqlMap["MYSQL_Tbl3dsRequest_Xxxxx"] = MYSQL_Tbl3dsRequest_Xxxxx
}
