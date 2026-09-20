//go:build db

package test

import (
	"context"
	"testing"
	"time"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/repository/model"
	"github.com/shopspring/decimal"

	// gormibmdb 内部通过 sql.Open("go_ibm_db", dsn) 建立连接，需注册该驱动
	_ "github.com/ibmdb/go_ibm_db"
)

func TestFindByStruct(t *testing.T) {
	ctx := context.Background()
	tmTeacherDao := GetRepository()

	testCases := []struct {
		name    string
		entity  *model.TmTeacher
		wantErr bool
	}{
		{name: "场景1:按主键查询(Id=242)", entity: &model.TmTeacher{Id: 242}, wantErr: false},
		{name: "场景2:单索引列-Name", entity: &model.TmTeacher{Name: "test1"}, wantErr: false},
		{name: "场景3:单索引列-ClassId", entity: &model.TmTeacher{ClassId: "1"}, wantErr: false},
		{name: "场景4:单索引列-Phone", entity: &model.TmTeacher{Phone: "13800000000"}, wantErr: false},
		{name: "场景5:联合索引-Name+Address", entity: &model.TmTeacher{Name: "test1", Address: "上海市浦东新区"}, wantErr: false},
		{name: "场景6:索引+普通列-Name+Salary", entity: &model.TmTeacher{Name: "test1", Salary: decimal.NewFromFloat(1000.0)}, wantErr: false},
		{name: "场景7:索引+特殊列-JpaVersion", entity: &model.TmTeacher{Name: "test1", JpaVersion: 1}, wantErr: false},
		{name: "场景8:索引+特殊列-CreateTime", entity: &model.TmTeacher{Name: "test1", CreateTime: time.Date(2026, 8, 13, 15, 35, 13, 775000000, time.UTC)}, wantErr: false},
		{name: "场景9:索引+特殊列-LastUpdateTime", entity: &model.TmTeacher{Name: "test1", LastUpdateTime: time.Date(2026, 8, 13, 15, 35, 13, 775000000, time.UTC)}, wantErr: false},
		{name: "场景10:仅特殊列无索引-预期错误", entity: &model.TmTeacher{JpaVersion: 1}, wantErr: true},
		{name: "场景11:空结构体-预期错误", entity: &model.TmTeacher{}, wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			res, err := tmTeacherDao.FindByStruct(ctx, tc.entity)
			if tc.wantErr {
				if err == nil {
					t.Errorf("期望错误但返回nil")
				} else {
					t.Logf("正确返回错误: %v", err)
				}
				return
			}
			if err != nil {
				t.Errorf("返回错误: %v", err)
				return
			}
			t.Logf("查询到 %d 条记录", len(res))
			for _, e := range res {
				t.Logf("结果: %+v", e)
			}
		})
	}
}


func printList[T any](name string, list []T, err error, t *testing.T) {

	t.Logf(" *** %v *** ", name)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	if len(list) == 0 {
		t.Log("未查询到预期的数据")
		t.Fail()
	}
	for _, entity := range list {
		t.Logf("结果:%v \n", entity)
	}
	t.Logf(" *** %v end *** ", name)
}

func printEntity(name string, entity interface{}, err error, t *testing.T) {
	t.Logf(" *** %v *** ", name)
	if err != nil {
		t.Log(err.Error())
		t.Fail()
	}
	t.Logf("查询结果%v \n", entity)
	t.Logf(" *** %v end *** ", name)
}
