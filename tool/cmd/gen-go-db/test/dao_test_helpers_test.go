package test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/dao"
	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

var (
	testDBOnce sync.Once
	testDB     *gorm.DB
)

// getTestDB 返回全局复用的测试数据库连接，避免每个用例重复建立连接池
func getTestDB() *gorm.DB {
	testDBOnce.Do(func() {
		testDB = mustOpenDB()
	})
	return testDB
}

// assertErrorContains 断言 err 非空且错误信息包含指定子串
func assertErrorContains(t *testing.T, err error, substr string) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误包含 %q，但得到 nil", substr)
	}
	if !strings.Contains(err.Error(), substr) {
		t.Fatalf("期望错误包含 %q，实际错误: %v", substr, err)
	}
}

// ============ 测试表自备数据（测试表策略） ============
// 策略：测试开始前清空整表并插入自备数据，测试结束后清空整表，
// 保证测试周期内表内数据完全干净可控，精确条数断言不依赖外部种子数据，
// 也不受表内其他数据影响。
//
// 注意：自备数据必须满足表约束——
//   - 主键唯一：同一测试内插入的自备数据主键不能重复
//   - 唯一约束列：tm_teacher 表的 card_no、phone 为唯一约束列（yaml constraints），
//     同一测试内插入的数据这两列不能重复；tbl_3ds_request 的
//     (RETRIEVAL_REFERENCE_NUMBER, MERCHANT_ID, TRANSACTION_TYPE) 为联合唯一约束
//   - 列长度：如 tm_teacher/tm_student 的 class_id 为 varchar(2)，插入值不能超长

// fixedCreateTime tm_student 自备数据中 NO001 的创建/更新时间（与 findstruct 场景8/9 断言一致）
var fixedCreateTime = time.Date(2026, 8, 13, 15, 35, 14, 0, time.Local)

// otherTime 其余自备数据行的创建/更新时间（统一非零值，避免插入零日期失败）
var otherTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local)

// tmStudentSeeds tm_student 测试表自备数据，满足 findstruct/condition/primaryKey 全部场景：
//   - Name=Alice 3 行(NO001/NO004/NO010)、ClassId=C01 2 行(NO001/NO010)
//   - Phone=13800000000 2 行(NO001/NO008)、CardNo=BJ001 2 行(NO001/NO008)
//   - TenantId=1 共 4 行(NO001/NO006/NO008/NO010)、其中 Alice 2 行
//   - TestBool 2 行(NO011 毕业/NO012 未毕业)、TestTime 2 行(NO014 入读2024/NO015 入读2023)
//   - TestZero 2 行、AllIdx/NO999、Helen 1 行、NO007(JpaVersion=99)
var tmStudentSeeds = []*model.TmStudent{
	{TenantId: 1, StudentNo: "NO001", Name: "Alice", Address: "北京市海淀区", Phone: "13800000000", ClassId: "C01", CardNo: "BJ001", Score: decimal.NewFromFloat(95.0), JpaVersion: 1, CreateTime: fixedCreateTime, LastUpdateTime: fixedCreateTime},
	{TenantId: 1, StudentNo: "NO010", Name: "Alice", Address: "上海市浦东新区", Phone: "13800000001", ClassId: "C01", CardNo: "BJ002", Score: decimal.NewFromFloat(90.0), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 2, StudentNo: "NO004", Name: "Alice", Address: "上海市徐汇区", Phone: "13800000002", ClassId: "C02", CardNo: "BJ003", Score: decimal.NewFromFloat(80.0), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO006", Name: "Eve", Address: "北京市朝阳区", Phone: "13800000003", ClassId: "C02", CardNo: "BJ004", CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO008", Name: "Dave", Address: "深圳市南山区", Phone: "13800000000", ClassId: "C02", CardNo: "BJ001", CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 4, StudentNo: "NO007", Name: "Fay", Address: "广州市天河区", Phone: "13800000004", ClassId: "C03", CardNo: "BJ005", JpaVersion: 99, CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 3, StudentNo: "NO013", Name: "Helen", Address: "杭州市西湖区", Phone: "13800000007", ClassId: "C04", CardNo: "BJ006", CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 5, StudentNo: "NO011", Name: "TestBool", Address: "测试市朝阳区", Phone: "13800000005", ClassId: "C05", CardNo: "BJ007", IsGraduate: true, CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 6, StudentNo: "NO012", Name: "TestBool", Address: "测试市海淀区", Phone: "13800000006", ClassId: "C05", CardNo: "BJ008", CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 7, StudentNo: "NO014", Name: "TestTime", Address: "测试市朝阳区", Phone: "13800000008", ClassId: "C06", CardNo: "BJ009", EnrollDate: timePtr(time.Date(2024, 9, 11, 0, 0, 0, 0, time.Local)), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 8, StudentNo: "NO015", Name: "TestTime", Address: "测试市海淀区", Phone: "13800000009", ClassId: "C06", CardNo: "BJ010", EnrollDate: timePtr(time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 9, StudentNo: "NO016", Name: "TestZero", Address: "测试市朝阳区", Phone: "13800000010", ClassId: "C07", CardNo: "BJ011", CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 10, StudentNo: "NO017", Name: "TestZero", Address: "测试市海淀区", Phone: "13800000011", ClassId: "C07", CardNo: "BJ012", CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 99, StudentNo: "NO999", Name: "AllIdx", Address: "测试市", Phone: "13800000999", ClassId: "C99", CardNo: "CQALL", CreateTime: otherTime, LastUpdateTime: otherTime},
}

// tmNoIndexSeeds tm_no_index 测试表自备数据（有主键无索引：tenant_id=1 共 4 行）
var tmNoIndexSeeds = []*model.TmNoIndex{
	{TenantId: 1, StudentNo: "NO001", Name: "A1", Address: "addr1", Phone: "13800000001", ClassId: "C1", CardNo: "CARD1", Score: decimal.NewFromInt(10), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO002", Name: "A2", Address: "addr2", Phone: "13800000002", ClassId: "C1", CardNo: "CARD2", Score: decimal.NewFromInt(20), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO003", Name: "A3", Address: "addr3", Phone: "13800000003", ClassId: "C1", CardNo: "CARD3", Score: decimal.NewFromInt(30), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO004", Name: "A4", Address: "addr4", Phone: "13800000004", ClassId: "C1", CardNo: "CARD4", Score: decimal.NewFromInt(40), CreateTime: otherTime, LastUpdateTime: otherTime},
}

// tmNoPrimarySeeds tm_no_primary 测试表自备数据（无主键有索引）：
//   - Name=Alice 3 行、其中 Address=北京市海淀区 1 行、Score=95 1 行
//   - ClassId=C01 2 行、Helen/13800000007 1 行
var tmNoPrimarySeeds = []*model.TmNoPrimary{
	{TenantId: 1, StudentNo: "NO001", Name: "Alice", Address: "北京市海淀区", Phone: "13800000000", ClassId: "C01", CardNo: "BJ001", Score: decimal.NewFromFloat(95.0), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO002", Name: "Alice", Address: "上海市浦东新区", Phone: "13800000001", ClassId: "C02", CardNo: "BJ002", Score: decimal.NewFromFloat(88.0), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 1, StudentNo: "NO003", Name: "Alice", Address: "广州市天河区", Phone: "13800000002", ClassId: "C01", CardNo: "BJ003", Score: decimal.NewFromFloat(80.0), CreateTime: otherTime, LastUpdateTime: otherTime},
	{TenantId: 2, StudentNo: "NO004", Name: "Helen", Address: "杭州市西湖区", Phone: "13800000007", ClassId: "C03", CardNo: "BJ004", CreateTime: otherTime, LastUpdateTime: otherTime},
}

// clearTable 清空指定测试表整表数据
func clearTable(t *testing.T, table string) {
	t.Helper()
	if err := getTestDB().Exec("DELETE FROM " + table).Error; err != nil {
		t.Fatalf("清空表 %s 失败: %v", table, err)
	}
}

// seedTmStudent 清空 tm_student 整表并插入自备数据，返回测试结束时的清理函数
func seedTmStudent(t *testing.T, ctx context.Context, studentDao dao.ITmStudentDao) func() {
	t.Helper()
	clearTable(t, "tm_student")
	for _, s := range tmStudentSeeds {
		if _, err := studentDao.InsertOne(ctx, s); err != nil {
			t.Fatalf("插入 tm_student 自备数据失败: %v", err)
		}
	}
	return func() { clearTable(t, "tm_student") }
}

// seedTmNoIndex 清空 tm_no_index 整表并插入自备数据，返回测试结束时的清理函数
func seedTmNoIndex(t *testing.T, ctx context.Context, tmNoIndexDao dao.ITmNoIndexDao) func() {
	t.Helper()
	clearTable(t, "tm_no_index")
	for _, s := range tmNoIndexSeeds {
		if _, err := tmNoIndexDao.InsertOne(ctx, s); err != nil {
			t.Fatalf("插入 tm_no_index 自备数据失败: %v", err)
		}
	}
	return func() { clearTable(t, "tm_no_index") }
}

// seedTmNoPrimary 清空 tm_no_primary 整表并插入自备数据，返回测试结束时的清理函数
func seedTmNoPrimary(t *testing.T, ctx context.Context, tmNoPrimaryDao dao.ITmNoPrimaryDao) func() {
	t.Helper()
	clearTable(t, "tm_no_primary")
	for _, s := range tmNoPrimarySeeds {
		if _, err := tmNoPrimaryDao.InsertOne(ctx, s); err != nil {
			t.Fatalf("插入 tm_no_primary 自备数据失败: %v", err)
		}
	}
	return func() { clearTable(t, "tm_no_primary") }
}
