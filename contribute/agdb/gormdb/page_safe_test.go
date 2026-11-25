package gormdb

import "testing"

// TestOrderStringRejectsUnsafe 覆盖排序列名与方向的校验：非法项返回空串，由 ToSqlOrder 丢弃。
func TestOrderStringRejectsUnsafe(t *testing.T) {
	cases := []struct {
		name  string
		order Order
		want  string
		why   string
	}{
		{"正常升序", Order{ColName: "id", Sort: ASC}, "id ASC", "合法列名与方向原样输出"},
		{"小写方向归一", Order{ColName: "id", Sort: OrderSort("asc")}, "id ASC", "方向大小写不敏感，统一为大写"},
		{"非法列名", Order{ColName: "id; drop", Sort: ASC}, "", "列名含分号必须丢弃"},
		{"表别名写法", Order{ColName: "t.id", Sort: ASC}, "t.id ASC", "点号用于 table.column，合法"},
		{"反引号写法", Order{ColName: "`order`", Sort: DESC}, "`order` DESC", "MySQL 保留字列的写法"},
		{"中文列名", Order{ColName: "订单号", Sort: ASC}, "", "只接受 ASCII，非 ASCII 丢弃"},
		{"函数列", Order{ColName: "DATE(created_at)", Sort: ASC}, "", "函数表达式不走排序字段"},
		{"方向为空", Order{ColName: "id"}, "", "空方向会生成非法 SQL，丢弃"},
		{"方向非法", Order{ColName: "id", Sort: OrderSort("asc; drop")}, "", "方向只允许 ASC/DESC"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.order.String(); got != c.want {
				t.Errorf("Order.String()=%q, want %q（%s）", got, c.want, c.why)
			}
		})
	}
}

// TestToSqlOrderDropsInvalid 验证多个排序项中非法项被丢弃且不产生空槽。
func TestToSqlOrderDropsInvalid(t *testing.T) {
	orders := []Order{
		{ColName: "id", Sort: ASC},
		{ColName: "bad;--", Sort: DESC},
		{ColName: "created_at", Sort: DESC},
	}
	want := "id ASC, created_at DESC"
	if got := ToSqlOrder(orders); got != want {
		t.Fatalf("ToSqlOrder()=%q, want %q（非法项应被丢弃，不产生空槽）", got, want)
	}
}

// TestOrderBuilderBuildDropsUnsafe 走 OrderBuilder.Build 端到端，确认链式 API 也被覆盖。
func TestOrderBuilderBuildDropsUnsafe(t *testing.T) {
	got := NewOrderBuilder().Asc("id").Desc("bad;--").Build()
	want := "id ASC"
	if got != want {
		t.Fatalf("OrderBuilder.Build()=%q, want %q（非法排序项应被丢弃）", got, want)
	}
}

// TestOrderBuilderAcceptsValid 守住绿侧：合法排序保持原样输出。
func TestOrderBuilderAcceptsValid(t *testing.T) {
	got := NewOrderBuilder().Asc("id").Desc("created_at").Build()
	want := "id ASC, created_at DESC"
	if got != want {
		t.Fatalf("OrderBuilder.Build()=%q, want %q", got, want)
	}
}
