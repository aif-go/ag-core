package excel

import (
	"fmt"
	"strconv"
	"strings"

	"ag-core/tool/cmd/gen-go-db/gendb/render"

	"github.com/xuri/excelize/v2"
)

// 其余模板文件转移为Excel模板文件
func OtherToExcel(config *render.AGInfraStructrueConfig, list []*render.TableData) {
	// 应该先判断文件是否存在,不存在才创建xlxs文件
	f := excelize.NewFile()
	defer f.Close()
	// 内容样式：蓝色字体、边框、左对齐
	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Color:  "0000FF",
			Size:   11,
			Bold:   true,
			Family: "宋体",
		},
		Border: []excelize.Border{
			{Type: "left", Style: 1, Color: "000000"},
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "right", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 1, Color: "000000"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			// Color: "0000FF",
			Size: 11,
			// Bold: true,
			Family: "宋体",
		},
		Border: []excelize.Border{
			{Type: "left", Style: 1, Color: "000000"},
			{Type: "top", Style: 1, Color: "000000"},
			{Type: "right", Style: 1, Color: "000000"},
			{Type: "bottom", Style: 1, Color: "000000"},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})

	f.SetActiveSheet(0)
	// var wg sync.WaitGroup
	// wg.Add(len(list))
	supportTablesLen := len(config.SupportTables)
	for _, data := range list {
		if _, ok := config.SupportTables[data.TableName]; !ok && supportTablesLen != 0 {
			// 设置了支持的表清单并且又不在支持的表清单的表被剔除
			continue
		}
		sheetName := data.TableName
		// log.Info("开始创建",sheetName,"的sheet")
		f.NewSheet(sheetName)
		// go func (sheetName string, data *render.TableData)  {

		// defer wg.Done()
		// defer f.SetColWidth(sheetName,"A","A",30)
		f.SetColWidth(sheetName, "A", "A", 30)
		// 构建表配置行数据
		f.SetCellValue(sheetName, "A1", "表存储引擎")
		f.SetCellValue(sheetName, "B1", "排序规则")
		f.SetCellValue(sheetName, "C1", "建表编码")
		f.SetCellValue(sheetName, "A2", data.Engine)
		f.SetCellValue(sheetName, "B2", data.Sort)
		f.SetCellValue(sheetName, "C2", data.Encode)
		f.SetCellStyle(sheetName, "A1", "C1", titleStyle)

		// 设置列数据 这个要根据excel的内容遍历循环写入第3行之后的内容,记录行号
		f.SetCellValue(sheetName, "A3", "列名")
		f.SetCellValue(sheetName, "B3", "列类型")
		f.SetCellValue(sheetName, "C3", "长度")
		f.SetCellValue(sheetName, "D3", "非空")
		f.SetCellValue(sheetName, "E3", "默认值")
		f.SetCellValue(sheetName, "F3", "主键")
		f.SetCellValue(sheetName, "G3", "自动增长")
		f.SetCellValue(sheetName, "H3", "描述")
		f.SetCellValue(sheetName, "I3", "自定义类型")
		// f.SetSheetRow(sheetName, "A3", []string{"列名","列类型","长度","非空","默认值","主键","自动增长","描述","自定义类型"})
		f.SetCellStyle(sheetName, "A3", "I3", titleStyle)

		rownum := 4
		// 开始写入每列的内容
		cell := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I"}
		rownumstr := ""
		// 画表结构部分，画出每个表
		for _, col := range data.ColumnList {
			rownumstr = strconv.Itoa(rownum)
			f.SetCellStr(sheetName, strings.Join([]string{"A", rownumstr}, ""), col.DbColName)
			f.SetCellStr(sheetName, strings.Join([]string{"B", rownumstr}, ""), col.GoType)
			if col.GoType == "time.Time" {
				f.SetCellStr(sheetName, strings.Join([]string{"B", rownumstr}, ""), "time")
			}
			if col.Length != "null" {
				f.SetCellStr(sheetName, strings.Join([]string{"C", rownumstr}, ""), col.Length)
			} else {
				// 对于日期类型，此处需要给一个长度在转换数据库ddl的时候，标识类型
				if strings.ToLower(col.DbType) == "date" {
					f.SetCellStr(sheetName, strings.Join([]string{"C", rownumstr}, ""), "8")
				}
			}
			if col.NotNullFlag {
				f.SetCellStr(sheetName, strings.Join([]string{"D", rownumstr}, ""), "Y")
			}
			f.SetCellStr(sheetName, strings.Join([]string{"E", rownumstr}, ""), col.DefaultVal)
			if col.PrimaryKey {
				f.SetCellStr(sheetName, strings.Join([]string{"F", rownumstr}, ""), "Y")
			}
			if col.AutoIncrement {
				f.SetCellStr(sheetName, strings.Join([]string{"G", rownumstr}, ""), "Y")
			}
			f.SetCellStr(sheetName, strings.Join([]string{"H", rownumstr}, ""), col.Comment)
			f.SetCellStr(sheetName, strings.Join([]string{"I", rownumstr}, ""), col.Description)
			f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"I", rownumstr}, ""), dataStyle)
			rownum = rownum + 1
		}
		// 开始写入
		rownum = rownum + 1
		rownumstr = strconv.Itoa(rownum)
		//主键
		f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), "主键")
		f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"A", rownumstr}, ""), titleStyle)
		// PRIMARY_KEY  横向展开
		rownum = rownum + 1
		rownumstr = strconv.Itoa(rownum)
		f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), "PRIMARY_KEY")
		f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"A", rownumstr}, ""), dataStyle)
		// f.SetSheetRow(sheetName,strings.Join([]string{"B",rownumstr},""),data.PrimaryKeyList)

		for index, primary := range strings.Split(data.PrimaryKeyList, ",") {
			f.SetCellValue(sheetName, strings.Join([]string{cell[index+1], rownumstr}, ""), primary)
			f.SetCellStyle(sheetName, strings.Join([]string{cell[index+1], rownumstr}, ""), strings.Join([]string{cell[index+1], rownumstr}, ""), dataStyle)
		}

		// 约束
		rownum = rownum + 2
		rownumstr = strconv.Itoa(rownum)
		f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), "约束")
		f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"B", rownumstr}, ""), titleStyle)

		// 约束的名字  列横向展开
		for _, unqiue := range data.UniqueIndexList {
			rownum = rownum + 1
			rownumstr = strconv.Itoa(rownum)
			f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), unqiue.IndexName)
			f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"B", rownumstr}, ""), dataStyle)

			for index, bindParam := range unqiue.BindParamList {
				f.SetCellValue(sheetName, strings.Join([]string{cell[index+1], rownumstr}, ""), bindParam.DbColName)
				f.SetCellStyle(sheetName, strings.Join([]string{cell[index+1], rownumstr}, ""), strings.Join([]string{cell[index+1], rownumstr}, ""), dataStyle)
			}
		}

		// 索引
		rownum = rownum + 2
		rownumstr = strconv.Itoa(rownum)
		f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), "索引")
		f.SetCellValue(sheetName, strings.Join([]string{"B", rownumstr}, ""), "索引类型")
		f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"B", rownumstr}, ""), titleStyle)
		// 索引名,BTREE,列...
		for _, general := range data.GeneralIndexList {
			rownum = rownum + 1
			rownumstr = strconv.Itoa(rownum)
			f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), general.IndexName)
			f.SetCellValue(sheetName, strings.Join([]string{"B", rownumstr}, ""), general.IndexType)
			f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"B", rownumstr}, ""), dataStyle)
			// f.SetSheetRow(sheetName,strings.Join([]string{"C",rownumstr},""),general.IndexColList)
			for index, bindParam := range general.BindParamList {
				f.SetCellValue(sheetName, strings.Join([]string{cell[index+2], rownumstr}, ""), bindParam.DbColName)
				f.SetCellStyle(sheetName, strings.Join([]string{cell[index+2], rownumstr}, ""), strings.Join([]string{cell[index+2], rownumstr}, ""), dataStyle)
			}
		}

		// 命名sql
		rownum = rownum + 2
		rownumstr = strconv.Itoa(rownum)
		f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), "自定义脚本名字")
		f.SetCellValue(sheetName, strings.Join([]string{"B", rownumstr}, ""), "脚本内容")
		f.SetCellValue(sheetName, strings.Join([]string{"C", rownumstr}, ""), "数据库类型")
		f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"C", rownumstr}, ""), titleStyle)
		// 方法名,sql内容,数据库类型
		for _, namingsql := range data.RNamingSqlList {
			rownum = rownum + 1
			rownumstr = strconv.Itoa(rownum)
			f.SetCellValue(sheetName, strings.Join([]string{"A", rownumstr}, ""), namingsql.MethodName)
			f.SetCellValue(sheetName, strings.Join([]string{"B", rownumstr}, ""), namingsql.NamingSql)
			f.SetSheetRow(sheetName, strings.Join([]string{"C", rownumstr}, ""), "")
			f.SetCellStyle(sheetName, strings.Join([]string{"A", rownumstr}, ""), strings.Join([]string{"C", rownumstr}, ""), dataStyle)
		}
		// }(sheetName,data)
	}
	// wg.Wait()
	fmt.Println("所有数据已经写入对应的sheet文档")
	f.DeleteSheet("Sheet1")
	f.SaveAs(config.OutputPath + "default.xlsx")
}
