package render

import (
	"log"
	"strconv"
	"strings"
)

var DBGoTypeReflect map[string]string

func init() {
	DBGoTypeReflect = map[string]string{}
	DBGoTypeReflect["varchar"] = "string"
	DBGoTypeReflect["varchar(n)"] = "string"
	DBGoTypeReflect["varchar2(n)"] = "string"
	DBGoTypeReflect["varchar2"] = "string"
	DBGoTypeReflect["timestamp"] = "time.Time"
	DBGoTypeReflect["datetime"] = "time.Time"
	DBGoTypeReflect["decimal(p)"] = "int64"
	DBGoTypeReflect["decimal(p,s)"] = "int64"
	DBGoTypeReflect["character(n)"] = "string"
	DBGoTypeReflect["character"] = "string"
	DBGoTypeReflect["char(n)"] = "string"
	DBGoTypeReflect["char"] = "string"
	DBGoTypeReflect["integer"] = "int32"
	DBGoTypeReflect["int"] = "int32"
	DBGoTypeReflect["bigint"] = "int64"
	DBGoTypeReflect["date"] = "time.Time"
}

// 将dbtype转换为go语言类型
func DbTypeConvertGoType(dbType string, decimal string) string {
	switch dbType {
	case "decimal(p)":
		return "int64"
	case "decimal": // 为excel输入的类型做准备
		fallthrough
	case "decimal(p,s)":
		if decimal == "0" || decimal == "" {
			return "int64"
		}
		return "float64"
	default:
		return DBGoTypeReflect[dbType]
	}
}

// 需要将特定的数据类型转换为ddl语句的格式
func FormtDbTypeToDDL(col *ColumnData) string {
	dbType := col.DbType
	length := col.Length
	switch dbType {
	case "varchar(n)":
		return "varchar(" + length + ")"
	case "character(n)":
		return "char(" + length + ")"
	case "decimal(p,s)":
		return "decimal(" + length + "," + col.DecimalLength + ")"
	case "decimal(p)":
		return "decimal(" + length + ")"
	default:
		return dbType
	}

}

// Imports 准备Imports的内容
var importFilterMap map[string][]string = map[string][]string{}

func Imports(orgGoType string, comment string) (string, string) {
	if strings.Contains(orgGoType, "time") {
		return "time.Time", "time"
	}

	if _, ok := importFilterMap[comment]; ok {
		return importFilterMap[comment][0], importFilterMap[comment][1]
	}

	if strings.Contains(comment, "///@Enum ") {
		///@Enum xxx/xxx/xxx.AAA-->xxx/xxx/xxx/AAA--->xxx/xxx/xxx  xxx/AAA
		newGoTypePck := strings.Replace(comment, "///@Enum ", "", -1)
		// index:=strings.LastIndex(newGoTypePck, "/")
		pck := strings.Split(newGoTypePck, "/")
		pkg := []string{}
		newGoType := []string{}
		length := len(pck)
		for i, v := range pck {
			if i >= length-2 {
				newGoType = append(newGoType, v)
			}
			if i < length-1 {
				pkg = append(pkg, v)
			}
		}
		importFilterMap[comment] = []string{strings.Join(newGoType, "."), strings.Join(pkg, "/")}
		return strings.Join(newGoType, "."), strings.Join(pkg, "/")
	}
	return orgGoType, ""
}

func ConvertDbTypeToGoType(dbType string, length string) string {
	switch dbType {
	case "numeric":
		// 穿透到下一个case处理,复用代码
		fallthrough
	case "decimal":
		if strings.Contains(length, ",") {
			decimalArr := strings.Split(length, ",")
			return DbTypeConvertGoType(dbType, decimalArr[1])
		} else {
			return DbTypeConvertGoType(dbType, "")
		}
	default:
		return DbTypeConvertGoType(dbType, "")
	}
}

// ConvertGoTypeToDbType 将Go类型转换为db类型
func ConvertGoTypeToDbType(goType string, length string, dbType string) string {
	switch goType {
	case "string":
		numlen, _ := strconv.Atoi(length)
		if numlen > 1 {
			return "VARCHAR(" + length + ")"
		}
		return "CHAR(" + length + ")"
	case "time":
		if length == "" {
			if dbType == "MYSQL" {
				return "datetime"
			}
			if dbType == "DB2" {
				return "timestamp"
			}
			// 默认返回datetime
			return "datetime"
		} else {
			return "date"
		}
	case "float64":
		return "decimal(" + length + ")"
	case "int32":
		return "int"
	case "int":
		return "int"
	case "int64":
		if length == "" {
			return "bigint"
		}
		return "decimal(" + length + ")"
	default:
		log.Printf("警告: %s类型暂时不支持转换为db类型，使用默认类型VARCHAR(255)", goType)
		return "VARCHAR(255)"
	}
}
