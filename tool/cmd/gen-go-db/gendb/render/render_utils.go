package render

import "strings"

// var entityImportFilterMap map[string]string = make(map[string]string,10)

// TableModel 构建entity的模板数据
func CreateTableModel(tableData *TableData, columnData *ColumnData) {

	if tableData.TableModelList == nil {
		tableData.TableModelList = []*TableModel{}
	}

	if tableData.Imports == nil {
		tableData.Imports = []string{}
	}

	// if _, ok := filterRepeatTypeMap[columnData.GoType]; !ok {
	goType, pck := Imports(columnData.GoType, "")
	if pck != "" {
		if _, ok := tableData.EntityImportsFilterMap.LoadOrStore(pck, pck); !ok {
			// map中不存在的场景才需要放入
			tableData.Imports = append(tableData.Imports, pck)
			// entityImportFilterMap[pck] = pck
		}
		columnData.GoType = goType
	}
	// 目前只要time需要额外的添加
	// if strings.Contains(columnData.GoType, "time") {
	// 	// map中不存在的场景才需要放入
	// 	filterRepeatTypeMap["time"] = "time"
	// 	tableData.Imports = append(tableData.Imports, "time")
	// }
	// 对于自定义的枚举类，根据标注引入的对应的包
	// TODO dao里面的imports怎么办?
	// }

	tableData.TableModelList = append(tableData.TableModelList, &TableModel{
		GoType:    columnData.GoType,
		GoColName: columnData.GoColName,
		GoTag:     CreateTag(columnData),
		DbColName: columnData.DbColName,
	})
}

// CreateTag 构建struct的tag标签
func CreateTag(columnData *ColumnData) string {

	builder := strings.Builder{}

	builder.WriteString(`gorm:"`)
	if columnData.AutoCreate {
		builder.WriteString("AUTOCREATETIME;")
	}
	if columnData.AutoUpdate {
		builder.WriteString("AUTOUPDATETIME;")
	}
	if columnData.JPAVersion {
		builder.WriteString("JPAVERSION;")
	}
	builder.WriteString(`column:`)

	builder.WriteString(columnData.DbColName)
	if columnData.DefaultVal != "" {
		builder.WriteString(";default:")
		builder.WriteString(columnData.DefaultVal)
	}
	if columnData.PrimaryKey {
		builder.WriteString(";primaryKey")
	}

	if columnData.NotNullFlag {
		builder.WriteString(";not null")
	}

	// TODO 如果一列数据被使用到多个
	if columnData.ColumnRefIndexList != nil {
		for _, ref := range columnData.ColumnRefIndexList {
			if ref == nil {
				continue
			}
			switch ref.IndexType {
			case General:
				builder.WriteString(";index:")

			case Unique:
				builder.WriteString(";uniqueIndex:")
			default:
				continue
			}
			builder.WriteString(ref.IndexName)
			if ref.Priority != "" {
				builder.WriteString(",priority:")
				builder.WriteString(ref.Priority)
			}
		}
	}

	// TODO 一个列可能被使用到多个索引上 此处应该是数组的机制
	// if columnData.GeneralIndexName != "" {
	// 	builder.WriteString(";index:")
	// 	builder.WriteString(columnData.GeneralIndexName)
	// 	if columnData.Priority != "" {
	// 		builder.WriteString(",priority:")
	// 		builder.WriteString(columnData.Priority)
	// 	}
	// }

	builder.WriteString(`"`)
	builder.WriteString(` json:"`)
	builder.WriteString(columnData.GoColName)
	builder.WriteString(`"`)

	return builder.String()
}
