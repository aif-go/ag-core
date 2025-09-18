package types

// MethodInfo 方法信息.
type MethodInfo struct {
	PkgInfo       `json:"pkg_info"` // 包信息，和上文一致，冗余
	ServiceName   string            `json:"service_name,omitempty"`    // 服务名，和上文一致，冗余
	Name          string            `json:"name,omitempty"`            // 方法名
	RawName       string            `json:"raw_name,omitempty"`        // 原始定义 m.desc.name()
	Oneway        bool              `json:"oneway,omitempty"`          // FIXME 暂无使用 原含义thrift中是否为oneway方法
	Void          bool              `json:"void,omitempty"`            // FIXME 暂无使用 proto中无void定义，thrift中使用
	Args          []*Parameter      `json:"args,omitempty"`            // 参数信息
	ArgsLength    int               `json:"args_length,omitempty"`     // 参数长度 len(Args) TODO 当前只允许1个参数
	Resp          *Parameter        `json:"resp,omitempty"`            // 响应信息
	Exceptions    []*Parameter      `json:"exceptions,omitempty"`      // FIXME 暂无使用 异常信息， proto中无定义
	ArgStructName string            `json:"arg_struct_name,omitempty"` // 参数实体名 MethodName + "Args"   目前只有kitex service生成使用
	ResStructName string            `json:"res_struct_name,omitempty"` // 响应实体名 MethodName + "Result" 目前只有kitex service生成使用

	IsResponseNeedRedirect bool `json:"is_response_need_redirect,omitempty"` // int -> int* // FIXME thrift使用

	GenArgResultStruct bool `json:"gen_arg_result_struct,omitempty"` // 是否生成参数实体类和响应实体类，proto解析时默认给true
	IsStreaming        bool `json:"is_streaming,omitempty"`          // 是否流式方法 : ClientStreaming || ServerStreaming
	ClientStreaming    bool `json:"client_streaming,omitempty"`      // 是否客户端流式方法
	ServerStreaming    bool `json:"server_streaming,omitempty"`      // 是否服务端流式方法

	// ext by houzw
	Deprected bool              // 是否废弃
	Comment   CommentSet        // 注释
	HttpDescs []*MethodHttpDesc // http描述，一个函数可能会被绑定多个http规则
}

// StreamingMode 返回方法的Stream类型，主要给模板识别使用
func (m *MethodInfo) StreamingMode() string {
	if !m.IsStreaming {
		return ""
	}
	if m.ClientStreaming {
		if m.ServerStreaming {
			return "bidirectional"
		}
		return "client"
	}
	if m.ServerStreaming {
		return "server"
	}
	return "unary"
}
