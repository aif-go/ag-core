package model

import (
	"testing"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// TestParseWhereExpr CHG-08 表驱动测试：where 表达式解析全形态矩阵。
func TestParseWhereExpr(t *testing.T) {
	cases := []struct {
		name string
		expr string
		want []table.WhereColField
	}{
		// 基础操作符
		{name: "等值", expr: "name = @Name", want: []table.WhereColField{{ColName: "name", FieldName: "Name", Operator: "="}}},
		{name: "不等于", expr: "status != @Status", want: []table.WhereColField{{ColName: "status", FieldName: "Status", Operator: "!="}}},
		{name: "大于", expr: "age > @Age", want: []table.WhereColField{{ColName: "age", FieldName: "Age", Operator: ">"}}},
		{name: "小于", expr: "age < @Age", want: []table.WhereColField{{ColName: "age", FieldName: "Age", Operator: "<"}}},
		{name: "大于等于", expr: "salary >= @Min", want: []table.WhereColField{{ColName: "salary", FieldName: "Min", Operator: ">="}}},
		{name: "小于等于", expr: "salary <= @Max", want: []table.WhereColField{{ColName: "salary", FieldName: "Max", Operator: "<="}}},
		// in 族
		{name: "in 参数", expr: "BIZ_DATE in @BizDateSlice", want: []table.WhereColField{{ColName: "BIZ_DATE", FieldName: "BizDateSlice", IsSlice: true, Operator: "in"}}},
		{name: "in 带括号参数", expr: "status in (@StatusList)", want: []table.WhereColField{{ColName: "status", FieldName: "StatusList", IsSlice: true, Operator: "in"}}},
		{name: "not in 参数", expr: "NAME not in @Names", want: []table.WhereColField{{ColName: "NAME", FieldName: "Names", IsSlice: true, Operator: "not in"}}},
		{name: "not in 带括号", expr: "NAME not in (@Names)", want: []table.WhereColField{{ColName: "NAME", FieldName: "Names", IsSlice: true, Operator: "not in"}}},
		// like 族
		{name: "like 参数带通配符", expr: "name like %_@Keyword%_", want: []table.WhereColField{{ColName: "name", FieldName: "Keyword", Operator: "like"}}},
		{name: "not like 参数", expr: "name not like @Keyword", want: []table.WhereColField{{ColName: "name", FieldName: "Keyword", Operator: "not like"}}},
		// between 双参数
		{name: "between 双参数", expr: "age between @Min AND @Max", want: []table.WhereColField{
			{ColName: "age", FieldName: "Min", Operator: "between"},
			{ColName: "age", FieldName: "Max", Operator: "between"},
		}},
		{name: "between 小写 and", expr: "age between @Min and @Max", want: []table.WhereColField{
			{ColName: "age", FieldName: "Min", Operator: "between"},
			{ColName: "age", FieldName: "Max", Operator: "between"},
		}},
		// 子串误判回归（CHG-08 核心修复点）
		{name: "参数名含 in 子串", expr: "salary >= @Min", want: []table.WhereColField{{ColName: "salary", FieldName: "Min", Operator: ">="}}},
		{name: "参数名含 in 子串2", expr: "age > @Index", want: []table.WhereColField{{ColName: "age", FieldName: "Index", Operator: ">"}}},
		{name: "列名含 in 子串", expr: "PRINT_DATE in @Dates", want: []table.WhereColField{{ColName: "PRINT_DATE", FieldName: "Dates", IsSlice: true, Operator: "in"}}},
		{name: "列名含 in 子串2", expr: "MINDATE in @Dates", want: []table.WhereColField{{ColName: "MINDATE", FieldName: "Dates", IsSlice: true, Operator: "in"}}},
		{name: "between 参数含 in", expr: "ts between @Begin AND @End", want: []table.WhereColField{
			{ColName: "ts", FieldName: "Begin", Operator: "between"},
			{ColName: "ts", FieldName: "End", Operator: "between"},
		}},
		// 无参数形态：条件由 conditonwhere 原样透传，不生成 Arg 字段
		{name: "is null", expr: "status IS NULL", want: nil},
		{name: "is not null", expr: "status IS NOT NULL", want: nil},
		{name: "between 字面量", expr: "age BETWEEN 18 AND 30", want: nil},
		{name: "等值字面量", expr: "deleted = 0", want: nil},
		{name: "子查询", expr: "id IN (SELECT id FROM t WHERE a > 1 AND b < 2)", want: nil},
		{name: "括号分组单表达式（超出范围，丢弃）", expr: "(a = 1 OR b = 2) AND c = @C", want: nil},
	}
	for _, tc := range cases {
		t.Run(tc.expr, func(t *testing.T) {
			got := parseWhereExpr(tc.expr)
			if len(got) != len(tc.want) {
				t.Fatalf("条目数不符: want %d got %d (%+v)", len(tc.want), len(got), got)
			}
			for i := range got {
				g, w := got[i], tc.want[i]
				if g.ColName != w.ColName || g.FieldName != w.FieldName || g.IsSlice != w.IsSlice || g.Operator != w.Operator {
					t.Fatalf("第%d条不匹配:\n  want: %+v\n  got:  %+v", i+1, w, g)
				}
			}
		})
	}
}
