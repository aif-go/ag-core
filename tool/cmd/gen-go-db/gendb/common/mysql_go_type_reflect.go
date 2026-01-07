package common

// MySQLGoTypeReflect 映射 MySQL 数据类型到 Go 类型
var MySQLGoTypeReflect = map[string]string{
	// 字符串类型
	"CHAR":       "string",
	"VARCHAR":    "string",
	"TINYTEXT":   "string",
	"TEXT":       "string",
	"MEDIUMTEXT": "string",
	"LONGTEXT":   "string",

	// 数值整型
	"TINYINT":    "int8",
	"BOOL":       "bool",
	"BOOLEAN":    "bool",
	"SMALLINT":   "int16",
	"MEDIUMINT":  "int32",
	"INT":        "int",
	"INTEGER":    "int",
	"BIGINT":     "int64",

	// 数值浮点型
	"FLOAT":      "float32",
	"DOUBLE":     "float64",
	"DECIMAL":    "*decimal.Decimal",
	"NUMERIC":    "*decimal.Decimal",

	// 日期时间类型
	"DATE":       "time.Time",
	"TIME":       "time.Time",
	"DATETIME":   "time.Time",
	"TIMESTAMP":  "time.Time",
	"YEAR":       "int16",

	// 二进制类型
	"BINARY":     "[]byte",
	"VARBINARY":  "[]byte",
	"TINYBLOB":   "[]byte",
	"BLOB":       "[]byte",
	"MEDIUMBLOB": "[]byte",
	"LONGBLOB":   "[]byte",

	// JSON类型
	"JSON":       "string",
}