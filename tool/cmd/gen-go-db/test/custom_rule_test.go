package test

import (
	"github.com/aif-go/ag-core/contribute/agdb/gormdb"
	"fmt"
	"testing"
	"time"
)


var tmpl=`SELECT a.* FROM (
    (SELECT 1 AS TABLE_ID, q.* , p.SETTLEMENT_AMOUNT,p.SETTLEMENT_DATE,p.SETTLEMENT_CURRENCY,p.AUTHORIZATION_CODE,p.RESPONSE_CODE,p.RESULT_CODE,p.TRANSACTION_IDENTIFIER 
     FROM (
        SELECT q.INSERT_TIMESTAMP, q.TRANSACTION_TYPE,q.ORDER_ID,q.PAN,q.TERMINAL_ID,q.MERCHANT_ID,q.TRANSMISSION_DATETIME,q.TRANSACTION_CURRENCY,q.TRANSACTION_AMOUNT,q.RETRIEVAL_REFERENCE_NUMBER, q.STAN,q.CARD_TYPE,q.ORDER_CURRENCY,q.ORDER_AMOUNT,q.DCC_FLAG,q.TDS_FLAG,null as REFUND_SEQUENCE, q.BILLING_AMOUNT,q.BILLING_CURRENCY 
        FROM TBL_PURCHASE_REQUEST q
        WHERE 1=1
        {{if .TransactionType}}
            AND q.TRANSACTION_TYPE = @TransactionType
        {{end}}
        {{if and (not .TransactionType) .TransactionTypeList (ne (len .TransactionTypeList) 0)}}
            AND q.TRANSACTION_TYPE IN (@TransactionTypeList)
        {{end}}
        {{if and .PAN (ne .PAN "")}}
            AND q.PAN = @PAN
        {{end}}
        {{if and .TerminalId (ne .TerminalId "")}}
            AND q.TERMINAL_ID = @TerminalId
        {{end}}
        {{if and .OrderId (ne .OrderId "")}}
            AND q.ORDER_ID = @OrderId
        {{end}}
        {{if and .RetrievalReferenceNumber (ne .RetrievalReferenceNumber "")}}
            AND q.RETRIEVAL_REFERENCE_NUMBER = @RetrievalReferenceNumber
        {{end}}
        {{if .BeginDateTime}}
            AND q.INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .EndDateTime}}
            AND q.INSERT_TIMESTAMP <= @EndDateTime
        {{end}}
        {{if and .MerchantId (ne .MerchantId "")}}
            AND q.MERCHANT_ID = @MerchantId
        {{end}}
    ) as q 
    LEFT JOIN (
        SELECT SETTLEMENT_AMOUNT,SETTLEMENT_DATE,SETTLEMENT_CURRENCY,AUTHORIZATION_CODE,RESPONSE_CODE,RESULT_CODE,TRANSACTION_IDENTIFIER,STAN,TRANSMISSION_DATETIME,TRANSACTION_TYPE 
        FROM TBL_PURCHASE_RESPONSE
        WHERE 1=1
        {{if and .ResponseCode (ne .ResponseCode "")}}
            AND p.RESPONSE_CODE = @ResponseCode
        {{end}}
        {{if .BeginDateTime}}
            AND INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .ResponseEndDateTime}}
            AND INSERT_TIMESTAMP <= @ResponseEndDateTime
        {{end}}
    ) p on q.TRANSMISSION_DATETIME = p.TRANSMISSION_DATETIME and q.STAN = p.STAN and q.TRANSACTION_TYPE = p.TRANSACTION_TYPE
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        INNER JOIN (
            SELECT m.MERCHANT_ID 
            FROM TBL_MERCHANT m 
            LEFT JOIN TBL_INSTITUTION i on i.INSTITUTION_CODE=m.INSTITUTION_CODE
            WHERE 1=1
            {{if and .BranchNo (ne .BranchNo "")}}
                AND (i.BRANCH_NO=@BranchNo or m.BRANCH_NO=@BranchNo)
            {{end}}
            {{if and .InstitutionCode (ne .InstitutionCode "")}}
                AND i.INSTITUTION_CODE=@InstitutionCode
            {{end}}
            {{if and .MerchantId (ne .MerchantId "")}}
                AND m.MERCHANT_ID=@MerchantId
            {{end}}
        ) as me on me.MERCHANT_ID =q.MERCHANT_ID
    {{end}}
    WHERE 1=1
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        AND me.MERCHANT_ID =q.MERCHANT_ID
    {{end}}
    {{if and .ResponseCode (ne .ResponseCode "")}}
        AND p.RESPONSE_CODE = @ResponseCode
    {{end}}
) 
UNION ALL (
    SELECT 2 AS TABLE_ID,q.*, p.SETTLEMENT_AMOUNT,p.SETTLEMENT_DATE,p.SETTLEMENT_CURRENCY,p.AUTHORIZATION_CODE,p.RESPONSE_CODE,p.RESULT_CODE,p.TRANSACTION_IDENTIFIER 
    FROM (
        SELECT q.INSERT_TIMESTAMP, q.TRANSACTION_TYPE,q.ORDER_ID,q.PAN,q.TERMINAL_ID,q.MERCHANT_ID,q.TRANSMISSION_DATETIME,q.TRANSACTION_CURRENCY,q.TRANSACTION_AMOUNT,q.RETRIEVAL_REFERENCE_NUMBER, q.STAN,q.CARD_TYPE,q.ORDER_CURRENCY,q.ORDER_AMOUNT,q.DCC_FLAG,q.TDS_FLAG,q.REFUND_SEQUENCE, q.BILLING_AMOUNT,q.BILLING_CURRENCY 
        FROM TBL_REFUND_REQUEST q
        WHERE 1=1
        {{if .TransactionType}}
            AND q.TRANSACTION_TYPE = @TransactionType
        {{end}}
        {{if and (not .TransactionType) .TransactionTypeList (ne (len .TransactionTypeList) 0)}}
            AND q.TRANSACTION_TYPE IN (@TransactionTypeList)
        {{end}}
        {{if and .PAN (ne .PAN "")}}
            AND q.PAN = @PAN
        {{end}}
        {{if and .TerminalId (ne .TerminalId "")}}
            AND q.TERMINAL_ID = @TerminalId
        {{end}}
        {{if and .OrderId (ne .OrderId "")}}
            AND q.ORDER_ID = @OrderId
        {{end}}
        {{if and .RetrievalReferenceNumber (ne .RetrievalReferenceNumber "")}}
            AND q.RETRIEVAL_REFERENCE_NUMBER = @RetrievalReferenceNumber
        {{end}}
        {{if .BeginDateTime}}
            AND q.INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .EndDateTime}}
            AND q.INSERT_TIMESTAMP <= @EndDateTime
        {{end}}
        {{if and .MerchantId (ne .MerchantId "")}}
            AND q.MERCHANT_ID = @MerchantId
        {{end}}
    ) as q 
    LEFT JOIN (
        SELECT SETTLEMENT_AMOUNT,SETTLEMENT_DATE,SETTLEMENT_CURRENCY,AUTHORIZATION_CODE,RESPONSE_CODE,RESULT_CODE,TRANSACTION_IDENTIFIER,TRANSMISSION_DATETIME,STAN,TRANSACTION_TYPE 
        FROM TBL_REFUND_RESPONSE
        WHERE 1=1
        {{if and .ResponseCode (ne .ResponseCode "")}}
            AND p.RESPONSE_CODE = @ResponseCode
        {{end}}
        {{if .BeginDateTime}}
            AND INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .ResponseEndDateTime}}
            AND INSERT_TIMESTAMP <= @ResponseEndDateTime
        {{end}}
    )p on q.TRANSMISSION_DATETIME = p.TRANSMISSION_DATETIME and q.STAN = p.STAN and q.TRANSACTION_TYPE = p.TRANSACTION_TYPE
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        INNER JOIN (
            SELECT m.MERCHANT_ID 
            FROM TBL_MERCHANT m 
            LEFT JOIN TBL_INSTITUTION i on i.INSTITUTION_CODE=m.INSTITUTION_CODE
            WHERE 1=1
            {{if and .BranchNo (ne .BranchNo "")}}
                AND (i.BRANCH_NO=@BranchNo or m.BRANCH_NO=@BranchNo)
            {{end}}
            {{if and .InstitutionCode (ne .InstitutionCode "")}}
                AND i.INSTITUTION_CODE=@InstitutionCode
            {{end}}
            {{if and .MerchantId (ne .MerchantId "")}}
                AND m.MERCHANT_ID=@MerchantId
            {{end}}
        ) as me on me.MERCHANT_ID =q.MERCHANT_ID
    {{end}}
    WHERE 1=1
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        AND me.MERCHANT_ID =q.MERCHANT_ID
    {{end}}
    {{if and .ResponseCode (ne .ResponseCode "")}}
        AND p.RESPONSE_CODE = @ResponseCode
    {{end}}
) 
UNION ALL (
    SELECT 3 AS TABLE_ID, q.*, p.SETTLEMENT_AMOUNT,p.SETTLEMENT_DATE,p.SETTLEMENT_CURRENCY,p.AUTHORIZATION_CODE,p.RESPONSE_CODE,p.RESULT_CODE,p.TRANSACTION_IDENTIFIER 
    FROM (
        SELECT q.INSERT_TIMESTAMP, q.TRANSACTION_TYPE,q.ORDER_ID,q.PAN,q.TERMINAL_ID,q.MERCHANT_ID,q.TRANSMISSION_DATETIME,q.TRANSACTION_CURRENCY,q.TRANSACTION_AMOUNT,q.RETRIEVAL_REFERENCE_NUMBER, q.STAN,q.CARD_TYPE,q.ORDER_CURRENCY,q.ORDER_AMOUNT,q.DCC_FLAG,q.TDS_FLAG,q.REFUND_SEQUENCE, q.BILLING_AMOUNT,q.BILLING_CURRENCY 
        FROM TBL_REVERSAL_REQUEST q
        WHERE 1=1
        {{if and .RetrievalReferenceNumber (ne .RetrievalReferenceNumber "")}}
            AND q.RETRIEVAL_REFERENCE_NUMBER =@RetrievalReferenceNumber
        {{end}}
        {{if .BeginDateTime}}
            AND q.INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .EndDateTime}}
            AND q.INSERT_TIMESTAMP <= @EndDateTime
        {{end}}
        {{if and .MerchantId (ne .MerchantId "")}}
            AND q.MERCHANT_ID =@MerchantId
        {{end}}
        {{if and .PAN (ne .PAN "")}}
            AND q.PAN = @PAN
        {{end}}
        {{if and .TerminalId (ne .TerminalId "")}}
            AND q.TERMINAL_ID =@TerminalId
        {{end}}
        {{if and .OrderId (ne .OrderId "")}}
            AND q.ORDER_ID =@OrderId
        {{end}}
        {{if .TransactionType}}
            AND q.TRANSACTION_TYPE = @TransactionType
        {{end}}
        {{if and (not .TransactionType) .TransactionTypeList (ne (len .TransactionTypeList) 0)}}
            AND q.TRANSACTION_TYPE IN (@TransactionTypeList)
        {{end}}
    ) as q 
    LEFT JOIN (
        SELECT SETTLEMENT_AMOUNT,SETTLEMENT_DATE,SETTLEMENT_CURRENCY,AUTHORIZATION_CODE,RESPONSE_CODE,RESULT_CODE,TRANSACTION_IDENTIFIER,TRANSMISSION_DATETIME,STAN ,TRANSACTION_TYPE 
        FROM TBL_REVERSAL_RESPONSE
        WHERE 1=1
        {{if and .ResponseCode (ne .ResponseCode "")}}
            AND p.RESPONSE_CODE =@ResponseCode
        {{end}}
        {{if .BeginDateTime}}
            AND INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .ResponseEndDateTime}}
            AND INSERT_TIMESTAMP <= @ResponseEndDateTime
        {{end}}
    ) p on q.TRANSMISSION_DATETIME = p.TRANSMISSION_DATETIME and q.STAN = p.STAN and q.TRANSACTION_TYPE = p.TRANSACTION_TYPE
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        INNER JOIN (
            SELECT m.MERCHANT_ID 
            FROM TBL_MERCHANT m 
            LEFT JOIN TBL_INSTITUTION i on i.INSTITUTION_CODE=m.INSTITUTION_CODE
            WHERE 1=1
            {{if and .BranchNo (ne .BranchNo "")}}
                AND (i.BRANCH_NO=@BranchNo or m.BRANCH_NO=@BranchNo)
            {{end}}
            {{if and .InstitutionCode (ne .InstitutionCode "")}}
                AND i.INSTITUTION_CODE=@InstitutionCode
            {{end}}
            {{if and .MerchantId (ne .MerchantId "")}}
                AND m.MERCHANT_ID=@MerchantId
            {{end}}
        ) as me on me.MERCHANT_ID =q.MERCHANT_ID
    {{end}}
    WHERE 1=1
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        AND me.MERCHANT_ID =q.MERCHANT_ID
    {{end}}
    {{if and .ResponseCode (ne .ResponseCode "")}}
        AND p.RESPONSE_CODE =@ResponseCode
    {{end}}
) 
UNION ALL (
    SELECT 4 AS TABLE_ID, d.* 
    FROM (
        SELECT d.INSERT_TIMESTAMP,d.TRANSACTION_TYPE,d.ORDER_ID,d.PAN,d.TERMINAL_ID,d.MERCHANT_ID,d.TRANSMISSION_DATETIME, d.TRANSACTION_CURRENCY, d.TRANSACTION_AMOUNT,d.RETRIEVAL_REFERENCE_NUMBER, d.STAN,d.CARD_TYPE,d.ORDER_CURRENCY,d.ORDER_AMOUNT,d.DCC_FLAG,d.TDS_FLAG,null as REFUND_SEQUENCE,d.BILLING_AMOUNT,d.BILLING_CURRENCY, null as SETTLEMENT_AMOUNT,null as SETTLEMENT_DATE,null as SETTLEMENT_CURRENCY,null as AUTHORIZATION_CODE,d.RESPONSE_CODE,d.RESULT_CODE,null as TRANSACTION_IDENTIFIER 
        FROM TBL_DCC_PREPURCHASE d
        WHERE 1=1
        {{if .TransactionType}}
            AND d.TRANSACTION_TYPE = @TransactionType
        {{end}}
        {{if and (not .TransactionType) .TransactionTypeList (ne (len .TransactionTypeList) 0)}}
            AND d.TRANSACTION_TYPE IN (@TransactionTypeList)
        {{end}}
        {{if and .PAN (ne .PAN "")}}
            AND d.PAN = @PAN
        {{end}}
        {{if and .TerminalId (ne .TerminalId "")}}
            AND d.TERMINAL_ID =@TerminalId
        {{end}}
        {{if and .OrderId (ne .OrderId "")}}
            AND d.ORDER_ID =@OrderId
        {{end}}
        {{if and .ResponseCode (ne .ResponseCode "")}}
            AND d.RESPONSE_CODE =@ResponseCode
        {{end}}
        {{if and .RetrievalReferenceNumber (ne .RetrievalReferenceNumber "")}}
            AND d.RETRIEVAL_REFERENCE_NUMBER =@RetrievalReferenceNumber
        {{end}}
        {{if .BeginDateTime}}
            AND d.INSERT_TIMESTAMP >= @BeginDateTime
        {{end}}
        {{if .EndDateTime}}
            AND d.INSERT_TIMESTAMP <= @EndDateTime
        {{end}}
        {{if and .MerchantId (ne .MerchantId "")}}
            AND d.MERCHANT_ID =@MerchantId
        {{end}}
    ) as d
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        INNER JOIN (
            SELECT m.MERCHANT_ID 
            FROM TBL_MERCHANT m 
            LEFT JOIN TBL_INSTITUTION i on i.INSTITUTION_CODE=m.INSTITUTION_CODE
            WHERE 1=1
            {{if and .BranchNo (ne .BranchNo "")}}
                AND (i.BRANCH_NO=@BranchNo or m.BRANCH_NO=@BranchNo)
            {{end}}
            {{if and .InstitutionCode (ne .InstitutionCode "")}}
                AND i.INSTITUTION_CODE=@InstitutionCode
            {{end}}
            {{if and .MerchantId (ne .MerchantId "")}}
                AND m.MERCHANT_ID=@MerchantId
            {{end}}
        ) as me on me.MERCHANT_ID =d.MERCHANT_ID
    {{end}}
    WHERE 1=1
    {{if and (or (and .BranchNo (ne .BranchNo "")) (and .InstitutionCode (ne .InstitutionCode ""))) (not (and .MerchantId (ne .MerchantId "")))}}
        AND me.MERCHANT_ID =d.MERCHANT_ID
    {{end}}
)`

type SQLParams struct {
	TransactionType        string      // 单个交易类型
	TransactionTypeList    []string    // 交易类型列表（IN条件）
	PAN                    string      // 卡号
	TerminalId             string      // 终端ID
	OrderId                string      // 订单ID
	RetrievalReferenceNumber string    // 检索参考号
	BeginDateTime          time.Time   // 开始时间
	EndDateTime            time.Time   // 结束时间
	MerchantId             string      // 商户ID
	ResponseCode           string      // 响应码
	ResponseEndDateTime    time.Time   // 响应结束时间
	BranchNo               string      // 分行号
	InstitutionCode        string      // 机构码
}
func BenchmarkTestFuckCustomRule(t *testing.B) {
	sqlParams:=&SQLParams{
		TransactionType: "1",
		PAN: "2",
		TerminalId: "123",
		OrderId: "456",
		RetrievalReferenceNumber: "1",
		BeginDateTime: time.Now(),
		EndDateTime: time.Now(),
		MerchantId: "1",
		ResponseCode: "000000",
		ResponseEndDateTime: time.Now(),
		BranchNo: "123",
		InstitutionCode: "000001",
	}
	sql, err:=gormdb.RendSql(tmpl,sqlParams)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	fmt.Println(sql)
}


// func rendSql(tmplStr string, params any) (string, error){
// 	// 步骤1：解析并渲染模板（得到带@属性名的SQL）
// 	tmpl, err := template.New("sql_tmpl").Funcs(template.FuncMap{
// 		"len": func(v []string) int { return len(v) },
// 	}).Parse(tmplStr)
// 	if err != nil {
// 		return "", fmt.Errorf("解析模板失败：%v", err)
// 	}
// 	var buf bytes.Buffer
// 	if err := tmpl.Execute(&buf, params); err != nil {
// 		return "", fmt.Errorf("渲染模板失败：%v", err)
// 	}
// 	rawSQL := buf.String()
// 	return rawSQL, nil
// }


// func init(){

// }


func TestTime(t *testing.T){
    var param time.Time
    if param.IsZero(){
        t.Log("结构体真的可以判断零值")
    }else{
        t.Log("时间类型无法通过此种方式判断零值")
    }

}
