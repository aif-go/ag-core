package protoc

import "testing"

// toHertzPath 是「Google API 路径参数 → Hertz 路由参数」的唯一转换点。
func TestToHertzPath(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"google 单个路径参数", "/api/cmdb/systems/{id}", "/api/cmdb/systems/:id"},
		{"google 多个路径参数", "/v1/{book}/{page}", "/v1/:book/:page"},
		{"带默认值", "/v1/systems/{page=2}", "/v1/systems/:page=2"},
		{"已是 hertz 写法时不变", "/api/cmdb/systems/:id", "/api/cmdb/systems/:id"},
		{"无路径参数时不变", "/api/cmdb/systems", "/api/cmdb/systems"},
		{"冒号分隔的默认值兼容", "/v1/{id:[0-9]+}", "/v1/:id=[0-9]+"},
		{"参数名含点和下划线", "/v1/{a.b_c}", "/v1/:a.b_c"},
	}
	for _, c := range cases {
		if got := toHertzPath(c.in); got != c.want {
			t.Errorf("%s: toHertzPath(%q) = %q, want %q", c.name, c.in, got, c.want)
		}
	}
}

// 回归：转换后的 path 必须能被既有解析逻辑识别，否则 {id} 会静默丢失参数。
func TestToHertzPathParamParity(t *testing.T) {
	path := toHertzPath("/api/cmdb/systems/{id}")

	vars := buildPathVars(path)
	if len(vars) != 1 {
		t.Fatalf("buildPathVars(%q) = %v, want 1 个参数", path, vars)
	}
	if _, ok := vars["id"]; !ok {
		t.Errorf("buildPathVars(%q) 缺少参数 id，实际 %v", path, vars)
	}

	names := extractParamNames(path)
	if len(names) != 1 || names[0] != "id" {
		t.Errorf("extractParamNames(%q) = %v, want [id]", path, names)
	}
}
