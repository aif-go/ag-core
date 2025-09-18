package types

type MethodHttpDesc struct {
	// method
	Name         string // 方法名
	OriginalName string // 原始定义 m.desc.name()
	Num          int    // 在同名函数中的序号，例如一个函数配置了多个http规则，则该函数的HttpDesc就有多个，需区分开
	Request      string // 入参类型名称 m.Input.GoIdent.GoName TODO 是否应该g.QualifiedGoIdent解析完整的引用名
	Reply        string // 响应类型名称 m.Output.GoIdent.GoName
	Comment      string // 注释

	// http_rule
	Path         string   // url 路径
	Method       string   // http请求类型 GET、POST等
	HasVars      bool     // 是否有路径参数
	PathVars     []string // url 路径参数
	HasBody      bool     // 是否有body参数
	Body         string   // body参数: * 代表整个请求对象 .Name 代表请求对象中 Name 属性
	ResponseBody string   // 响应body参数: * 代表整个响应对象，.Msg 代表http响应只返回响应对象的Msg属性
}
