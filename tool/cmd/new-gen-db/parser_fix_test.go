package main

import (
	"ag-core/tool/cmd/new-gen-db/excel"
	"fmt"
	"testing"
)

func TestParserFix(t *testing.T) {
	// 测试用例1: 用户提到的问题
	whereExpr := "ORDER_ID = @OrderId AND MERCHANT_ID = @MerchantId AND TRANSACTION_TYPE = @TransactionType AND RESPONSE_CODE = '00'"
	fmt.Printf("测试用例1: %s\n", whereExpr)
	result := excel.ParseWhereCondition(whereExpr)
	if result != nil {
		fmt.Printf("操作符: %s\n", result.Operator)
		fmt.Printf("条件数量: %d\n", len(result.Conditions))
		for i, cond := range result.Conditions {
			fmt.Printf("条件%d: %s\n", i+1, cond.Expr)
		}
	} else {
		fmt.Println("解析结果为空")
	}
	fmt.Println()
	
	// 测试用例2: 包含OR操作符
	whereExpr2 := "STATUS = 'ACTIVE' OR STATUS = 'PENDING'"
	fmt.Printf("测试用例2: %s\n", whereExpr2)
	result2 := excel.ParseWhereCondition(whereExpr2)
	if result2 != nil {
		fmt.Printf("操作符: %s\n", result2.Operator)
		fmt.Printf("条件数量: %d\n", len(result2.Conditions))
		for i, cond := range result2.Conditions {
			fmt.Printf("条件%d: %s\n", i+1, cond.Expr)
		}
	} else {
		fmt.Println("解析结果为空")
	}
	fmt.Println()
	
	// 测试用例3: 包含括号
	whereExpr3 := "(STATUS = 'ACTIVE' OR STATUS = 'PENDING') AND TYPE = 'USER'"
	fmt.Printf("测试用例3: %s\n", whereExpr3)
	result3 := excel.ParseWhereCondition(whereExpr3)
	if result3 != nil {
		fmt.Printf("操作符: %s\n", result3.Operator)
		fmt.Printf("条件数量: %d\n", len(result3.Conditions))
		for i, cond := range result3.Conditions {
			if cond.Expr != "" {
				fmt.Printf("条件%d: %s\n", i+1, cond.Expr)
			} else if len(cond.Conditions) > 0 {
				fmt.Printf("条件%d: 嵌套条件 - 操作符: %s, 子条件数: %d\n", i+1, cond.Operator, len(cond.Conditions))
				for j, subCond := range cond.Conditions {
					fmt.Printf("  子条件%d: %s\n", j+1, subCond.Expr)
				}
			}
		}
	} else {
		fmt.Println("解析结果为空")
	}
}