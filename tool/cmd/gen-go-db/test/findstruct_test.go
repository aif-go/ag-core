package test

import (
	"context"
	"testing"
	"time"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
)

func TestStudentFindByStruct(t *testing.T) {
	ctx := context.Background()
	studentDao := GetStudentRepository()

	// 固定时间，与数据库测试数据一致（DSN loc=Local，故用 time.Local）
	fixedCreateTime := time.Date(2026, 8, 13, 15, 35, 14, 0, time.Local)

	testCases := []struct {
		name    string
		entity  *model.TmStudent
		wantErr bool
		wantCnt int // 预期返回条数（-1 表示不检查）
	}{
		// ============ 1-9: 基础索引场景 ============
		{
			name:    "场景1:复合主键全字段-TenantId+StudentNo",
			entity:  &model.TmStudent{TenantId: 1, StudentNo: "NO001"},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景2:单索引-Name",
			entity:  &model.TmStudent{Name: "Alice"},
			wantErr: false,
			wantCnt: 3, // NO001, NO004, NO010
		},
		{
			name:    "场景3:单索引-ClassId",
			entity:  &model.TmStudent{ClassId: "C01"},
			wantErr: false,
			wantCnt: 2,
		},
		{
			name:    "场景4:单索引-Phone",
			entity:  &model.TmStudent{Phone: "13800000000"},
			wantErr: false,
			wantCnt: 2,
		},
		{
			name:    "场景5:联合索引-Name+Address",
			entity:  &model.TmStudent{Name: "Alice", Address: "北京市海淀区"},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景6:索引+普通列-Name+Score",
			entity:  &model.TmStudent{Name: "Alice", Score: 95.0},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景7:索引+特殊列-Name+JpaVersion",
			entity:  &model.TmStudent{Name: "Alice", JpaVersion: 1},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景8:索引+CreateTime",
			entity:  &model.TmStudent{Name: "Alice", CreateTime: fixedCreateTime},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景9:索引+LastUpdateTime",
			entity:  &model.TmStudent{Name: "Alice", LastUpdateTime: fixedCreateTime},
			wantErr: false,
			wantCnt: 1,
		},

		// ============ 10-11: 错误场景 ============
		{
			name:    "场景10:仅特殊列-无索引-预期错误",
			entity:  &model.TmStudent{JpaVersion: 1},
			wantErr: true,
			wantCnt: 0,
		},
		{
			name:    "场景11:空结构体-无索引-预期错误",
			entity:  &model.TmStudent{},
			wantErr: true,
			wantCnt: 0,
		},

		// ============ 复合主键进阶场景 ============
		{
			name:    "场景12:复合主键全字段+普通索引-三条件",
			entity:  &model.TmStudent{TenantId: 1, StudentNo: "NO006", Name: "Eve"},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景13:复合主键仅首键-TenantId-主键索引生效",
			entity:  &model.TmStudent{TenantId: 1},
			wantErr: false,
			wantCnt: 4, // 首键命中主键索引（最左前缀），返回 tenant_id=1 全部行
		},
		{
			name:    "场景14:复合主键部分+普通索引-TenantId+Name",
			entity:  &model.TmStudent{TenantId: 1, Name: "Alice"},
			wantErr: false,
			wantCnt: 2, // NO001(TenantId=1), NO010(TenantId=1, NO004 TenantId=2)
		},
		{
			name:    "场景15:复合主键全字段+仅特殊列",
			entity:  &model.TmStudent{TenantId: 4, StudentNo: "NO007", JpaVersion: 99},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景16:复合主键全零值边界+Name",
			entity:  &model.TmStudent{TenantId: 0, StudentNo: "", Name: "Alice"},
			wantErr: false,
			wantCnt: 3, // TenantId=0不进WHERE, 靠Name命中3条
		},
		{
			name:    "场景17:复合主键仅首键+普通列-主键索引生效",
			entity:  &model.TmStudent{TenantId: 1, Score: 100},
			wantErr: false,
			wantCnt: 0, // 首键命中主键索引；tenant_id=1 且 score=100 无匹配行
		},

		// ============ 其他覆盖场景 ============
		{
			name:    "场景18:单独CardNo索引",
			entity:  &model.TmStudent{CardNo: "BJ001"},
			wantErr: false,
			wantCnt: 2, // NO001, NO008
		},
		{
			name:    "场景19:多独立索引-Name+Phone",
			entity:  &model.TmStudent{Name: "Helen", Phone: "13800000007"},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景20:Address单独-联合索引非前导列-预期错误",
			entity:  &model.TmStudent{Address: "北京市海淀区"},
			wantErr: true,
			wantCnt: 0,
		},
		{
			name:    "场景21:BoolTest过滤",
			entity:  &model.TmStudent{Name: "TestBool", IsGraduate: true},
			wantErr: false,
			wantCnt: 1, // 仅NO011 is_graduate=1
		},
		{
			name:    "场景22:TimePointer过滤",
			entity:  &model.TmStudent{Name: "TestTime", EnrollDate: timePtr(time.Date(2024, 9, 11, 0, 0, 0, 0, time.Local))},
			wantErr: false,
			wantCnt: 1,
		},
		{
			name:    "场景23:Score=0.0+Name-零值double不进WHERE",
			entity:  &model.TmStudent{Name: "TestZero", Score: 0.0},
			wantErr: false,
			wantCnt: 2, // Score=0不进WHERE, Name命中2条
		},
		{
			name:    "场景24:全索引列同时设置",
			entity:  &model.TmStudent{
				TenantId:   99,
				StudentNo:  "NO999",
				Name:       "AllIdx",
				ClassId:    "C99",
				Phone:      "13800000999",
				CardNo:     "CQALL",
			},
			wantErr: false,
			wantCnt: 1,
		},
	}

	for _, tc := range testCases {
		tc := tc // 避免循环变量捕获
		t.Run(tc.name, func(t *testing.T) {
			res, err := studentDao.FindByStruct(ctx, tc.entity)

			// 期望错误
			if tc.wantErr {
				if err == nil {
					t.Errorf("[%s] 期望错误但返回nil, 结果: %v", tc.name, res)
				} else {
					t.Logf("[%s] 正确返回错误: %v", tc.name, err)
				}
				return
			}

			// 期望成功
			if err != nil {
				t.Errorf("[%s] 不期望错误但返回: %v", tc.name, err)
				return
			}

			// 校验条数
			if tc.wantCnt > 0 && len(res) != tc.wantCnt {
				t.Errorf("[%s] 期望 %d 条，实际 %d 条, 结果: %v", tc.name, tc.wantCnt, len(res), res)
				return
			}

			// 打印结果
			t.Logf("[%s] 查询到 %d 条记录", tc.name, len(res))
			for _, e := range res {
				t.Logf("  -> %+v", e)
			}
		})
	}
}



// === 辅助函数 ===

// timePtr 辅助函数：将 time.Time 转为 *time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}