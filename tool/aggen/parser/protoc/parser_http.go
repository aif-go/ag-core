/* 改造自 kratos 解析http规则 */
package protoc

import (
	"ag-core/tool/aggen/types"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"unicode"

	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var methodSets = make(map[string]int)

func parserMethodHttpDesc(m *protogen.Method) ([]*types.MethodHttpDesc, error) {
	hds := make([]*types.MethodHttpDesc, 0)

	// 获取http规则 option (google.api.http)
	rule, ok := proto.GetExtension(m.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
	if rule != nil && ok {
		// 解析http配置
		// 主规则
		hd, err := buildHTTPRule(
			// g,
			// service,
			m,
			rule,
			// omitemptyPrefix,
		)
		if err != nil {
			return nil, err
		}
		hds = append(hds, hd) // 主规则在第一个
		// 额外规则
		for _, bind := range rule.AdditionalBindings {
			hd, err := buildHTTPRule(m, bind)
			if err != nil {
				return nil, err
			}
			hds = append(hds, hd)
		}

	}

	// testLog("parserMethodHttpDesc", hds)

	return hds, nil
}

func buildHTTPRule(
	m *protogen.Method,
	rule *annotations.HttpRule,
) (*types.MethodHttpDesc, error) {
	var (
		path         string
		method       string
		body         string
		responseBody string
	)

	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = http.MethodGet
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = http.MethodPut
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = http.MethodPost
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = http.MethodDelete
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = http.MethodPatch
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	if method == "" {
		// method = http.MethodPost
		return nil, fmt.Errorf("http method is empty")
	}

	// 默认路径
	if path == "" {
		// TODO
		// path = fmt.Sprintf("%s/%s/%s", omitemptyPrefix, service.Desc.FullName(), m.Desc.Name())
		return nil, fmt.Errorf("default path is empty")
	}

	body = rule.Body
	responseBody = rule.ResponseBody

	// slog.Info("http_rule:", "method", method, "path", path, "body", body, "responseBody", responseBody)

	md, err := buildMethodDesc(
		// g,
		m,
		method,
		path,
	)
	if err != nil {
		return nil, err
	}
	// GET 和 DELETE 方法不允许有 body
	if method == http.MethodGet || method == http.MethodDelete {
		if body != "" {
			// _, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s body should not be declared.\n", method, path)
			return nil, fmt.Errorf("%s %s body should not be declared", method, path)
		}
	} else {
		if body == "" {
			// _, _ = fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: %s %s does not declare a body.\n", method, path)
			return nil, fmt.Errorf("%s %s does not declare a body", method, path)
		}
	}
	if body == "*" {
		md.HasBody = true
		md.Body = ""
	} else if body != "" {
		md.HasBody = true
		md.Body = "." + CamelCaseVars(body)
	} else {
		md.HasBody = false
	}
	if responseBody == "*" {
		md.ResponseBody = ""
	} else if responseBody != "" {
		md.ResponseBody = "." + CamelCaseVars(responseBody)
	}
	return md, nil
}

func buildMethodDesc(
	// g *protogen.GeneratedFile,
	m *protogen.Method,
	method,
	path string) (*types.MethodHttpDesc, error) {
	defer func() { methodSets[m.GoName]++ }()

	vars := buildPathVars(path)

	for v, s := range vars {
		fields := m.Input.Desc.Fields()
		// fmt.Printf("=======%s", fields)

		if s != nil {
			path = replacePath(v, *s, path)
		}
		for _, field := range strings.Split(v, ".") {
			if strings.TrimSpace(field) == "" {
				continue
			}
			if strings.Contains(field, ":") {
				field = strings.Split(field, ":")[0]
			}
			field = lowerFirst(field)
			fd := fields.ByName(protoreflect.Name(field))
			if fd == nil {
				return nil, fmt.Errorf("The corresponding field '%s' declaration in message could not be found in '%s'", v, path)
				// fmt.Fprintf(os.Stderr, "\u001B[31mERROR\u001B[m: The corresponding field '%s' declaration in message could not be found in '%s'\n", v, path)
				// os.Exit(2)
			}
			if fd.IsMap() {
				return nil, fmt.Errorf("The field in path:'%s' shouldn't be a map", v)
				// fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a map.\n", v)
			} else if fd.IsList() {
				return nil, fmt.Errorf("The field in path:'%s' shouldn't be a list", v)
				// fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: The field in path:'%s' shouldn't be a list.\n", v)
			} else if fd.Kind() == protoreflect.MessageKind || fd.Kind() == protoreflect.GroupKind {
				fields = fd.Message().Fields()
			}
		}
	}
	comment := m.Comments.Leading.String() + m.Comments.Trailing.String()
	if comment != "" {
		comment = "// " + m.GoName + strings.TrimPrefix(strings.TrimSuffix(comment, "\n"), "//")
	}

	pathVars := extractParamNames(path)
	return &types.MethodHttpDesc{
		Name:         m.GoName,
		OriginalName: string(m.Desc.Name()),
		Num:          methodSets[m.GoName],
		// TODO
		// Request:      g.QualifiedGoIdent(m.Input.GoIdent),
		Request: m.Input.GoIdent.GoName,
		// Reply:        g.QualifiedGoIdent(m.Output.GoIdent),
		Reply:    m.Output.GoIdent.GoName,
		Comment:  comment,
		Path:     path,
		PathVars: pathVars,
		Method:   method,
		HasVars:  len(vars) > 0,
	}, nil
}

func buildPathVars(path string) (res map[string]*string) {
	if strings.HasSuffix(path, "/") {
		fmt.Fprintf(os.Stderr, "\u001B[31mWARN\u001B[m: Path %s should not end with \"/\" \n", path)
	}
	pattern := regexp.MustCompile(`(?i):([\w.-]+)(?:=([^/]+))?`)
	matches := pattern.FindAllStringSubmatch(path, -1)
	res = make(map[string]*string, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(m[1])
		if len(name) > 1 && len(m[2]) > 0 {
			res[name] = &m[2]
		} else {
			res[name] = nil
		}
	}
	return
}

func replacePath(name string, value string, path string) string {
	pattern := regexp.MustCompile(fmt.Sprintf(`(?i):%s\b(?:=([^/]*))?`, regexp.QuoteMeta(name)))
	idx := pattern.FindStringIndex(path)
	if len(idx) > 0 {
		path = fmt.Sprintf("%s{%s:%s}%s",
			path[:idx[0]], // The start of the match
			name,
			strings.ReplaceAll(value, "*", ".*"),
			path[idx[1]:],
		)
	}
	return path
}

func extractParamNames(pattern string) []string {
	re := regexp.MustCompile(`:(\w+)`)
	matches := re.FindAllStringSubmatch(pattern, -1)

	params := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}
	return params
}
func lowerFirst(s string) string {
	rs := []rune(s)
	rs[0] = unicode.ToLower(rs[0])
	return string(rs)
}

func CamelCaseVars(s string) string {
	subs := strings.Split(s, ".")
	vars := make([]string, 0, len(subs))
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")
}

// camelCase returns the CamelCased name.
// If there is an interior underscore followed by a lower case letter,
// drop the underscore and convert the letter to upper case.
// There is a remote possibility of this rewrite causing a name collision,
// but it's so remote we're prepared to pretend it's nonexistent - since the
// C++ generator lowercase names, it's extremely unlikely to have two fields
// with different capitalization.
// In short, _my_field_name_2 becomes XMyFieldName_2.
func camelCase(s string) string {
	if s == "" {
		return ""
	}
	t := make([]byte, 0, 32)
	i := 0
	if s[0] == '_' {
		// Need a capital letter; drop the '_'.
		t = append(t, 'X')
		i++
	}
	// Invariant: if the next letter is lower case, it must be converted
	// to upper case.
	// That is, we process a word at a time, where words are marked by _ or
	// upper case letter. Digits are treated as words.
	for ; i < len(s); i++ {
		c := s[i]
		if c == '_' && i+1 < len(s) && isASCIILower(s[i+1]) {
			continue // Skip the underscore in s.
		}
		if isASCIIDigit(c) {
			t = append(t, c)
			continue
		}
		// Assume we have a letter now - if not, it's a bogus identifier.
		// The next word is a sequence of characters that must start upper case.
		if isASCIILower(c) {
			c ^= ' ' // Make it a capital letter.
		}
		t = append(t, c) // Guaranteed not lower case.
		// Accept lower case sequence that follows.
		for i+1 < len(s) && isASCIILower(s[i+1]) {
			i++
			t = append(t, s[i])
		}
	}
	return string(t)
}

// Is c an ASCII lower-case letter?
func isASCIILower(c byte) bool {
	return 'a' <= c && c <= 'z'
}

// Is c an ASCII digit?
func isASCIIDigit(c byte) bool {
	return '0' <= c && c <= '9'
}
