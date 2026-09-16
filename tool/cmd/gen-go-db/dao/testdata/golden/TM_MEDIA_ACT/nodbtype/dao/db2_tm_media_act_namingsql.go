package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

const DB2_TmMediaAct_NoPageQuery = "SELECT CARDNO,APP_ID,NEW_CARDNO FROM TM_MEDIA_ACT WHERE (BIZ_DATE = @BizDate AND ACTION_CD = @ActionCd)"

const DB2_TmMediaAct_Xxxxx = "SELECT CARDNO,APP_ID,NEW_CARDNO FROM (SELECT CARDNO,APP_ID,NEW_CARDNO, ROW_NUMBER() OVER(ORDER BY SEQ) AS RN  FROM TM_MEDIA_ACT WHERE ((BIZ_DATE = @BizDate AND ACTION_CD = @ActionCd) OR (ADDRESS = @Address AND BIZ_DATE in @BizDateSlice))) AS T WHERE RN BETWEEN @Start AND @End"

const DB2_TmMediaAct_Xxxxx_Count = "SELECT COUNT(*) FROM TM_MEDIA_ACT WHERE ((BIZ_DATE = @BizDate AND ACTION_CD = @ActionCd) OR (ADDRESS = @Address AND BIZ_DATE in @BizDateSlice))"

func InitTmMediaActDB2() {
	TmMediaActNamingSqlMap["DB2_TmMediaAct_NoPageQuery"] = DB2_TmMediaAct_NoPageQuery
	TmMediaActNamingSqlMap["DB2_TmMediaAct_Xxxxx"] = DB2_TmMediaAct_Xxxxx
	TmMediaActNamingSqlMap["DB2_TmMediaAct_Xxxxx_Count"] = DB2_TmMediaAct_Xxxxx_Count
}
