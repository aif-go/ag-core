package gendb

import (
	// "ag-core-inner-agdb/ag/ag_db/conditonwhere"
	"ag-core/tool/cmd/gen-go-db/gendb/render"
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	// "gorm.io/driver/mysql"
)

func TestExcelGen(t *testing.T) {
	//{{.Module}}
	moduleByte, _ := exec.Command("go", "list", "-f", "{{.Module.Path}}", ".").Output()
	packagePath:= strings.TrimSpace(string(moduleByte))
	fmt.Println("目标路径:" + packagePath)
	config := &render.AGInfraStructrueConfig{
		BaseConfig: render.BaseConfig{
			// DbTemplatePath: "C:/Users/songbing/Desktop/goerm/mps.erm",
			DbTemplatePath: "./repository/yaml/",
			// DbTemplatePath:    "./mps-template.xlsx",
			// DbTemplatePath:    "C:/Users/songbing/Desktop/generate/mps-template.xlsx",
			// DbTemplatePath: "C:/Users/songbing/Desktop/generate/repository/yaml/",
			PackageNamePrefix: packagePath,
			// OutputPath:        "C:/Users/songbing/Desktop/generate/",
			OutputPath: "./",
			DbType:            "",
		},
		GenerateOptions: render.GenerateOptions{
			Sqlable:    true,
			Daoable:    true,
			Entityable: true,
		},
		SupportConfig: render.SupportConfig{
			SupportDB: []string{"mysql", "db2"},
			SupportTables: map[string]string{
				// "tbl_3ds_request": "1",
			},
		},
	}
	err := GenerateDBGoFile(config)
	if err != nil {
		t.Log("生成失败:", err)
	}
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

func TestFileList(t *testing.T) {
	var yamlFiles []string
	// 检查 rootDir 是否为文件
	rootDir := "./tm_test.yaml"
	info, err := os.Stat(rootDir)
	if err != nil {
		t.Log(fmt.Errorf("访问路径失败: %w", err).Error())
		t.Fail()
		return
	}

	// 如果是文件，检查扩展名
	if !info.IsDir() {
		ext := strings.ToLower(filepath.Ext(rootDir))
		if ext == ".yaml" || ext == ".yml" {
			t.Log("是yaml文件")
			return
		}
	}

	// 如果是目录，遍历查找 YAML 文件
	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Log(fmt.Errorf("遍历文件失败: %w", err).Error())
			t.Fail()
			return err
		}

		if info.IsDir() {
			// 如果不递归且不是根目录，则跳过子目录
			if path != rootDir {
				return filepath.SkipDir
			}
			return nil
		}

		// 检查文件扩展名是否为 .yaml 或 .yml
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".yaml" || ext == ".yml" {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})

	if err != nil {
		t.Log(fmt.Errorf("遍历目录失败: %w", err).Error())
		t.Fail()
	}

	for _, yamlFile := range yamlFiles {
		t.Logf("找到 YAML 文件: %s", yamlFile)
	}

}


func TestSearchGoMod(t *testing.T) {
	// 获取当前程序执行目录
	currentDir, err := os.Getwd()
	if err != nil {
		t.Log(err)
		return
	}

	// 向上遍历目录，查找 go.mod
	for {
		goModPath := filepath.Join(currentDir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// 找到 go.mod，解析第一行 module 声明
			file, err := os.Open(goModPath)
			if err != nil {
				t.Log(err)
				return
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if strings.HasPrefix(line, "module ") {
					t.Logf("module name: %s", strings.TrimPrefix(line, "module "))
					return
				}
			}
		}

		// 到达根目录仍未找到
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			t.Log("未找到 go.mod 文件")
			return
		}
		currentDir = parentDir
	}
	

}
