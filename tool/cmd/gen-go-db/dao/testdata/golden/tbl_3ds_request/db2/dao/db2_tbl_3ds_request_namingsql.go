package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

const DB2_Tbl3dsRequest_Xxxxx = "SELECT * FROM tbl_3ds_request WHERE (RETRIEVAL_REFERENCE_NUMBER = @RRN)"

func InitTbl3dsRequestDB2() {
	Tbl3dsRequestNamingSqlMap["DB2_Tbl3dsRequest_Xxxxx"] = DB2_Tbl3dsRequest_Xxxxx
}
