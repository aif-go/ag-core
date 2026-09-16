package dao

// DO NOT EDIT
// DO NOT EDIT
// DO NOT EDIT

const MYSQL_TmMediaAct_NoPageQuery = "SELECT CARDNO,APP_ID,NEW_CARDNO FROM TM_MEDIA_ACT WHERE (BIZ_DATE = @BizDate AND ACTION_CD = @ActionCd)"

const MYSQL_TmMediaAct_Xxxxx = "SELECT CARDNO,APP_ID,NEW_CARDNO FROM TM_MEDIA_ACT WHERE ((BIZ_DATE = @BizDate AND ACTION_CD = @ActionCd) OR (ADDRESS = @Address AND BIZ_DATE in @BizDateSlice)) LIMIT @Start, @End"

const MYSQL_TmMediaAct_Xxxxx_Count = "SELECT COUNT(*) FROM TM_MEDIA_ACT WHERE ((BIZ_DATE = @BizDate AND ACTION_CD = @ActionCd) OR (ADDRESS = @Address AND BIZ_DATE in @BizDateSlice))"

func InitTmMediaActMYSQL() {
	TmMediaActNamingSqlMap["MYSQL_TmMediaAct_NoPageQuery"] = MYSQL_TmMediaAct_NoPageQuery
	TmMediaActNamingSqlMap["MYSQL_TmMediaAct_Xxxxx"] = MYSQL_TmMediaAct_Xxxxx
	TmMediaActNamingSqlMap["MYSQL_TmMediaAct_Xxxxx_Count"] = MYSQL_TmMediaAct_Xxxxx_Count
}
