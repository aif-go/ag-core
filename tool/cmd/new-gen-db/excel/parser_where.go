package excel

import (
	"strings"
)

// TokenType token类型
type TokenType int

const (
	TokenTypeExpr   TokenType = iota // 表达式
	TokenTypeAND                     // AND操作符
	TokenTypeOR                      // OR操作符
	TokenTypeLParen                  // 左括号
	TokenTypeRParen                  // 右括号
)

// Token token结构
type Token struct {
	Type  TokenType
	Value string
}

// Lexer 词法分析器
type Lexer struct {
	input string
	pos   int
}

// NewLexer 创建词法分析器
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: input,
		pos:   0,
	}
}

// NextToken 获取下一个token
func (l *Lexer) NextToken() *Token {
	// 跳过空白字符
	for l.pos < len(l.input) && (l.input[l.pos] == ' ' || l.input[l.pos] == '\t') {
		l.pos++
	}

	if l.pos >= len(l.input) {
		return nil
	}

	// 检查是否是括号
	if l.input[l.pos] == '(' {
		l.pos++
		return &Token{Type: TokenTypeLParen, Value: "("}
	}
	if l.input[l.pos] == ')' {
		l.pos++
		return &Token{Type: TokenTypeRParen, Value: ")"}
	}

	// 检查是否是操作符，需要确保前后都是空白字符或者是表达式边界
	// 检查AND操作符
	if l.pos+2 < len(l.input) && strings.ToUpper(l.input[l.pos:l.pos+3]) == "AND" {
		// 检查前面是否是表达式边界（字符串开始或空格）
		prevIsBoundary := l.pos == 0 || l.input[l.pos-1] == ' ' || l.input[l.pos-1] == '\t'
		// 检查后面是否是表达式边界（字符串结束或空格）
		nextIsBoundary := l.pos+3 == len(l.input) || l.input[l.pos+3] == ' ' || l.input[l.pos+3] == '\t' || l.input[l.pos+3] == '('
		
		if prevIsBoundary && nextIsBoundary {
			l.pos += 3
			return &Token{Type: TokenTypeAND, Value: "AND"}
		}
	}
	
	// 检查OR操作符
	if l.pos+1 < len(l.input) && strings.ToUpper(l.input[l.pos:l.pos+2]) == "OR" {
		// 检查前面是否是表达式边界（字符串开始或空格）
		prevIsBoundary := l.pos == 0 || l.input[l.pos-1] == ' ' || l.input[l.pos-1] == '\t'
		// 检查后面是否是表达式边界（字符串结束或空格）
		nextIsBoundary := l.pos+2 == len(l.input) || l.input[l.pos+2] == ' ' || l.input[l.pos+2] == '\t' || l.input[l.pos+2] == '('
		
		if prevIsBoundary && nextIsBoundary {
			l.pos += 2
			return &Token{Type: TokenTypeOR, Value: "OR"}
		}
	}

	// 否则是表达式，直到遇到操作符、括号或结束
	start := l.pos
	for l.pos < len(l.input) {
		// 检查是否是操作符的开始
		if l.pos+2 < len(l.input) && strings.ToUpper(l.input[l.pos:l.pos+3]) == "AND" {
			// 检查前面是否是表达式边界（字符串开始或空格）
			prevIsBoundary := l.pos == 0 || l.input[l.pos-1] == ' ' || l.input[l.pos-1] == '\t'
			// 检查后面是否是表达式边界（字符串结束或空格）
			nextIsBoundary := l.pos+3 == len(l.input) || l.input[l.pos+3] == ' ' || l.input[l.pos+3] == '\t' || l.input[l.pos+3] == '('
			
			if prevIsBoundary && nextIsBoundary {
				break
			}
		}
		if l.pos+1 < len(l.input) && strings.ToUpper(l.input[l.pos:l.pos+2]) == "OR" {
			// 检查前面是否是表达式边界（字符串开始或空格）
			prevIsBoundary := l.pos == 0 || l.input[l.pos-1] == ' ' || l.input[l.pos-1] == '\t'
			// 检查后面是否是表达式边界（字符串结束或空格）
			nextIsBoundary := l.pos+2 == len(l.input) || l.input[l.pos+2] == ' ' || l.input[l.pos+2] == '\t' || l.input[l.pos+2] == '('
			
			if prevIsBoundary && nextIsBoundary {
				break
			}
		}
		// 检查是否是括号
		if l.input[l.pos] == '(' || l.input[l.pos] == ')' {
			break
		}
		l.pos++
	}
	return &Token{Type: TokenTypeExpr, Value: strings.TrimSpace(l.input[start:l.pos])}
}

// Parser 语法分析器
type Parser struct {
	lexer *Lexer
	curr  *Token
}

// NewParser 创建语法分析器
func NewParser(lexer *Lexer) *Parser {
	return &Parser{
		lexer: lexer,
		curr:  lexer.NextToken(),
	}
}

// Parse 解析where条件
func (p *Parser) Parse() *WhereClause {
	return p.parseExpression()
}

// parseExpression 解析表达式
func (p *Parser) parseExpression() *WhereClause {
	clause := &WhereClause{
		Operator:   "AND", // 默认操作符
		Conditions: []*Condition{},
	}

	// 解析第一个条件
	cond := p.parseCondition()
	if cond != nil {
		clause.Conditions = append(clause.Conditions, cond)
	}

	// 解析后续的条件和操作符
	for p.curr != nil && (p.curr.Type == TokenTypeAND || p.curr.Type == TokenTypeOR) {
		// 记录操作符
		operator := p.curr.Value
		clause.Operator = operator
		p.curr = p.lexer.NextToken()

		// 解析下一个条件
		cond := p.parseCondition()
		if cond != nil {
			clause.Conditions = append(clause.Conditions, cond)
		}
	}

	return clause
}

// parseCondition 解析条件
func (p *Parser) parseCondition() *Condition {
	if p.curr == nil {
		return nil
	}

	// 处理括号
	if p.curr.Type == TokenTypeLParen {
		p.curr = p.lexer.NextToken()
		nestedClause := p.parseExpression()
		if p.curr != nil && p.curr.Type == TokenTypeRParen {
			p.curr = p.lexer.NextToken()
		}

		// 创建嵌套条件
		nestedConditions := make([]*Condition, len(nestedClause.Conditions))
		for i, cond := range nestedClause.Conditions {
			nestedConditions[i] = cond
		}

		return &Condition{
			Operator:   nestedClause.Operator,
			Conditions: nestedConditions,
		}
	}

	// 处理表达式
	if p.curr.Type == TokenTypeExpr {
		expr := p.curr.Value
		p.curr = p.lexer.NextToken()

		return &Condition{
			Expr: expr,
		}
	}

	return nil
}

// ParseWhereCondition 解析where条件
func ParseWhereCondition(whereExpr string) *WhereClause {
	// 去除首尾空白
	whereExpr = strings.TrimSpace(whereExpr)
	if whereExpr == "" {
		return nil
	}

	// 创建词法分析器和语法分析器
	lexer := NewLexer(whereExpr)
	parser := NewParser(lexer)

	// 解析where条件
	return parser.Parse()
}
