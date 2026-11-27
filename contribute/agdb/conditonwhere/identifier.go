package conditonwhere

import "strings"

// IsSafeIdentifier 判断字段名是否可安全拼进 SQL。
//
// 标识符无法参数化，只能在拼接前校验。规则是“放行合法形态、拒绝 SQL 元字符”：
// 允许 ASCII 字母、数字、下划线、点号（table.column），以及成对反引号包起来的
// 标识符（MySQL 保留字写法）；其余字符一律拒绝——空白、引号、分号、括号能切断
// 标识符，减号、斜杠、星号等是注释和运算符。
//
// 这里坚持“默认拒绝”而不是列坏字符的黑名单：漏判一个合法写法只是多拒一次，
// 漏判一个危险字符就是注入；白名单的失败方向是安全的。
//
// 只接受 ASCII：本项目列名全部由 gen-go-db 生成，实测 1589 个列名无一非 ASCII；
// 放开 Unicode 只会引入同形字和额外开销。确有非 ASCII 标识符时走命名 SQL 的
// 表达式路径，不要从这里开口子。
func IsSafeIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for {
		i := strings.IndexByte(name, '.')
		if i < 0 {
			return isSafeIdentifierSegment(name)
		}
		if !isSafeIdentifierSegment(name[:i]) {
			return false
		}
		name = name[i+1:]
	}
}

// isSafeIdentifierSegment 校验单个标识符段（不含点号）。
// 段可以用一对反引号包起来，此时允许数字开头——MySQL 里 `1col` 是合法的。
func isSafeIdentifierSegment(segment string) bool {
	if segment == "" {
		return false
	}
	quoted := false
	if len(segment) >= 2 && segment[0] == '`' && segment[len(segment)-1] == '`' {
		segment = segment[1 : len(segment)-1]
		quoted = true
		if segment == "" {
			return false
		}
	}
	for i := 0; i < len(segment); i++ {
		c := segment[i]
		switch {
		case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			continue
		case c >= '0' && c <= '9':
			// 未加引号时标识符不能以数字开头
			if !quoted && i == 0 {
				return false
			}
			continue
		default:
			return false
		}
	}
	return true
}
