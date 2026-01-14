package gendb

import (
	// "ag-core-inner-agdb/ag/ag_db/conditonwhere"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	// "gorm.io/driver/mysql"
)

func TestExcelGen(t *testing.T) {
	//{{.Module}}
	moduleByte, _ := exec.Command("go", "list", "-f", "", ".").Output()
	fmt.Println("目标路径:" + strings.TrimSpace(string(moduleByte)))
	config := &render.AGInfraStructrueConfig{
		BaseConfig: render.BaseConfig{
			// DbTemplatePath: "C:/Users/songbing/Desktop/goerm/mps.erm",
			DbTemplatePath: "./tm_test.yaml",
			// DbTemplatePath: "./mps-template.xlsx",
			PackageNamePrefix: "ag-core-inner-agdb/tool/aggen/gendb",
			OutputPath:        "./",
			DbType:            "mysql",
		},
		GenerateOptions: render.GenerateOptions{
			Sqlable:    true,
			Daoable:    true,
			Entityable: true,
		},
		SupportConfig: render.SupportConfig{
			SupportDB: []string{"mysql", "db2"},
			SupportTables: map[string]string{
				"tm_test": "TmTest",
			},
		},
	}
	GenerateDBGoFile(config)
	// GenerateYamlFile(config)
}

func TestInsert(t *testing.T) {

	// dao:=GetRepository()
	// ctx:=context.Background()
	// tmTest:=&model.TmTest{
	// 	// Seq: 1,
	// 	Sex: 0,
	// 	Name: "刘启",
	// 	Phone: "13800000007",
	// 	Address: "上海市浦东新区",
	// 	// JpaVersion: 2,
	// }
	// printEntity("InsertOne",count,err,t)
}

func TestInsertIgnoreZeroVal(t *testing.T) {
	// dao:=GetRepository()
	// ctx:=context.Background()
	// tmTest:=&model.TmTest{
	// 	Sex: 1,
	// 	Name: "十一",
	// 	Phone: "13800000011",
	// 	Address: "上海市浦东新区",
	// }
	// count,err:=dao.InsertOneIgnoreZeroVal(ctx,tmTest)
	// printEntity("InsertOneIgnoreZeroVal",count,err,t)
}
func BenchmarkChainWhereQuery(b *testing.B) {

	// var Cardno, BizDate, Address conditonwhere.IndexField = "CARDNO","BIZ_DATE","ADDRESS"
	// b.ResetTimer()
	// for i:=0;i<=b.N;i++{
	// 	// dao:=GetRepository()
	// 	// ctx:=context.Background()
	// 	builder:=conditonwhere.NewChainBuilder()
	// 	// 模板会定义一个IndexField，用于控制条件列
	// 	builder=builder.AND(conditonwhere.Gt(Cardno, "1"),conditonwhere.Eq(BizDate,time.Now())).OR(conditonwhere.And(conditonwhere.In(Cardno,"1","2","3"),conditonwhere.Neq(Address,1)))
	// 	builder.Build()
	// 	// fmt.Printf("SQL: %s\n", wheresql)
	// 	// fmt.Printf("Args: %v\n", args)
	// 	// dao.FindByCardnoBizDateWhere(ctx, builder)
	// }

}

// func TestQuery(t *testing.T){
// 	dao:=GetRepository()
// 	ctx:=context.Background()

// appidlist:= &model.FindByAppIdArg{
// 	AppIdSlice: []string{"1","2"},
// }

// list1,err:=dao.FindByAppId(ctx, appidlist)
// printList("FindByAppId",list1,err,t)
// arg := &model.FindBySelfColArg{

// }
// list2,err:=dao.FindBySelfCol(ctx,arg)
// printList("FindBySelfCol",list2,err,t)

// 	list,error:=dao.FindByAddressBizDate(ctx,"上海市浦东新区",time.Date(2025,time.October,11,0,0,0,0,time.Now().Location()))
// 	printList("FindByAddressBizDate", list, error, t)

// 	entity,er:=dao.FindByCardnoBizDate(ctx, "1", time.Date(2025,time.October,11,0,0,0,0,time.Now().Location()))
// 	printEntity("FindByIdNoIdType", entity, er, t)

// 	list1,er1:=dao.FindByBizDateActionCd(ctx,time.Date(2025,time.October,11,0,0,0,0,time.Now().Location()),"C")
// 	printList("FindByBizDateActionCd", list1, er1, t)

// 	entity1,er2:=dao.FindByPrimaryKey(ctx,1)
// 	printEntity("FindByPrimaryKey", entity1, er2, t)

// }

// func TestNamingSql(t *testing.T){
// 	dao:=GetRepository()
// 	ctx:=context.Background()

// 	arg:=&model.TmTestXxxxxArg{
// 		Phone: "13800000005",
// 	}
// 	res,err:=dao.Xxxxx(ctx, arg)
// 	printEntity("Xxxxx",res,err,t)
// }

// TestPageQuery 测试分页查询
func TestPageQuery(t *testing.T) {
	// dao:=GetRepository()
	// ctx:=context.Background()
	// res,err:=dao.FindAaByPage(ctx, &model.TmTestFindAaByPageArg{
	// 	Address: "上海市浦东新区",
	// 	PageNum: 2,
	// 	PageSize: 3,
	// })
	// printEntity("Xxxxx",res,err,t)
}

func TestPageFindAllCols(t *testing.T) {
	// dao:=GetRepository()
	// ctx:=context.Background()
	// res,err:=dao.FindAllColsByPage(ctx, &model.TmTestFindAllColsByPageArg{
	// 	Address: "上海市浦东新区",
	// 	PageNum: 2,
	// 	PageSize: 3,
	// })
	// printEntity("TestPageFindAllCols",res,err,t)
}

func TestFindAllCols(t *testing.T) {
	// dao:=GetRepository()
	// ctx:=context.Background()
	// res,err:=dao.FindAllCols(ctx, &model.TmTestFindAllColsArg{
	// 	Address: "上海市浦东新区",
	// })
	// printEntity("TestFindAllCols",res,err,t)
}

func TestUpdate(t *testing.T) {
	// dao:=GetRepository()
	// ctx:=context.Background()
	// test:=&model.TmTest{
	// 	Seq: 1,
	// 	// Sex: 3,
	// 	Name: "张三",
	// 	// Phone: "13800000000",
	// 	// Address: "上海市浦东新区",
	// 	JpaVersion: 2,
	// }

	// dao.UpdateByPriIngoreNullCols(ctx, test)
	// tmMediaAct:= &model.TmMediaAct{
	// 	// Seq: 3,
	// 	Cardno: "N1234567890",
	// 	AppId: "1",
	// 	NewCardno: "N0987654321",
	// 	BizDate: time.Now(),
	// 	ActionCd: "C",
	// 	ActStatus: "S",
	// 	IdNo: "4",
	// 	IdType: "A",
	// 	Address: "上海市浦东新区",
	// 	Org: "00000000",
	// 	// JpaVersion: 0,
	// 	// CreatedTime: time.Now(),
	// 	// LastModifiedTime: time.Now(),
	// }
	// count, error:=dao.UpdateByPrimaryKey(ctx, tmMediaAct)
	// printEntity("UpdateByPrimaryKey",count, error,t)

	// arg1:= &model.TmMediaAct{
	// 	Cardno: "M1234567890",
	// 	NewCardno: "M0987654321",
	// }
	// count1, error1:=dao.UpdateByIdNoIdType(ctx,"1","A",arg1);
	// printEntity("UpdateByIdNoIdType",count1, error1,t)

	// arg2:= &model.TmMediaAct{
	// 	Zip: "邮政编码",
	// }
	// bizdate:=time.Date(2026,time.October,13,0,0,0,0,time.Now().Location())
	// count2, error2:=dao.UpdateByBizDateActionCd(ctx,bizdate,"C",arg2)
	// printEntity("UpdateByBizDateActionCd",count2, error2,t)

	// arg3:= &model.TmMediaAct{
	// 	City: "河南省",
	// }
	// count3, error3:=dao.UpdateByAddressBizDate(ctx,"上海市浦东新区", bizdate, arg3)
	// printEntity("UpdateByBizDateActionCd",count3, error3,t)

	// dao.UpdateByPriIngoreNullCols()
}

// func TestDelete(t *testing.T){
// 	ctx:=context.Background()
// 	dao:=GetRepository()
// 	del1,err1:=dao.DeleteByPrimaryKey(ctx, 1)
// 	printEntity("DeleteByPrimaryKey",del1,err1,t)

// 	bizdate:=time.Date(2025,time.October,14,0,0,0,0,time.Now().Location())
// 	del2,err2:=dao.DeleteByAddressBizDate(ctx,"上海市浦东新区",bizdate)
// 	printEntity("DeleteByAddressBizDate",del2,err2,t)
// 	del3,err3:=dao.DeleteByBizDateActionCd(ctx,time.Date(2025,time.October,15,0,0,0,0,time.Now().Location()),"C")
// 	printEntity("DeleteByBizDateActionCd",del3,err3,t)

// 	del4,err4:=dao.DeleteByIdNoIdType(ctx,"5","A")
// 	printEntity("DeleteByIdNoIdType",del4,err4,t)

// }

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

// func GetRepository() dao.ITmTestDao {

// 	dsn := "root:root@tcp(localhost:3306)/process?parseTime=True&loc=Local"
// 	db,err:=gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err!=nil{
// 		panic(err.Error())
// 	}

// 	// 此处测试ok
// 	// res:=db.Raw("select * from tm_media_act where app_id in ?",[]string{"1" , "2"}).Find(&[]*model.TmMediaAct{})

// 	// fmt.Println(res.RowsAffected)

// 	sqldb, err:=db.DB()
// 	if err!=nil{
// 		panic(err.Error())
// 	}

// 	sqldb.SetMaxIdleConns(10)
// 	sqldb.SetMaxOpenConns(10)
// 	sqldb.SetConnMaxLifetime(time.Second)
// 	sqldb.SetConnMaxIdleTime(time.Second)

// 	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
// 		Level: slog.LevelDebug, // 输出级别：Debug及以上
// 	})
// 	logger := slog.New(jsonHandler)
// 	repository:=gormdb.NewRepository(logger,db)

// 	return dao.NewTmTestDao(repository)
// }
