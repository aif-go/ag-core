package table

import "strings"

// ZeroCheckExpr 返回列零值/非零值判断表达式——所有生成器的唯一事实源。
// owner 为变量名（如 "entity"、"tmTeacher"）；
// invert=false 生成"是否为零值"条件，true 生成"非零值"条件（GuardCheck 语义）。
// 新增 GoType 分支只需修改本函数。
func ZeroCheckExpr(owner string, col ColumnData, invert bool) string {
	field := owner + "." + col.JsonTag

	// 指针类型以 nil 为零值
	if strings.HasPrefix(col.GoType, "*") {
		if invert {
			return field + " != nil"
		}
		return field + " == nil"
	}

	switch col.GoType {
	case "string":
		if invert {
			return field + ` != ""`
		}
		return field + ` == ""`
	case "time.Time", "decimal.Decimal":
		if invert {
			return "!" + field + ".IsZero()"
		}
		return field + ".IsZero()"
	case "optimisticlock.Version":
		// 乐观锁列零值 = 未装载，以 Valid 判断（值 0 属合法历史版本）
		if invert {
			return field + ".Valid"
		}
		return "!" + field + ".Valid"
	case "bool":
		// bool 类型的零值为 false
		if invert {
			return field
		}
		return "!" + field
	default:
		// 数值类型（int, int64, float64 等）
		if invert {
			return field + " != 0"
		}
		return field + " == 0"
	}
}
