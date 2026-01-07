package common

// DB2GoTypeReflect 映射 DB2 数据类型到 Go 类型
var DB2GoTypeReflect = map[string]string{
	// 字符串类型
	"CHAR":        "string",  
	"VARCHAR":     "string",// 32,672字节
	"LONG VARCHAR": "string", // 32,700字节
	"CLOB":        "string",  // 2,147,483,647 字节
	"GRAPHIC":     "string", //  127 定长图形化 双字节
	"VARGRAPHIC":  "string",// 16,336 变长图形化 双字节
	"LONG VARGRAPHIC": "string",// 16,350 变长图形化 双字节
	"DBCLOB":      "string",

	// 数值类型
	"SMALLINT":    "int16",
	"INTEGER":     "int",
	"INT":         "int",
	"BIGINT":      "int64",
	"DECIMAL":     "*decimal.Decimal",
	"NUMERIC":     "*decimal.Decimal",
	"REAL":        "float32",
	"DOUBLE":      "float64",
	"DECFLOAT":    "float64",

	// 布尔类型
	"BOOLEAN":     "bool",

	// 日期时间类型
	"DATE":        "time.Time",
	"TIME":        "time.Time",
	"TIMESTAMP":   "time.Time",
	"TIMESTAMPTZ": "time.Time",

	// 二进制类型
	"BINARY":      "[]byte",
	"VARBINARY":   "[]byte",
	"LONG VARBINARY": "[]byte",
	"BLOB":        "[]byte",

	// 其他类型
	"XML":         "string",
}