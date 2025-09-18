package types

// ServiceInfo 服务信息.
type ServiceInfo struct {
	PkgInfo               // 包信息，和上文一致，冗余
	ServiceName    string // 服务名   - s.GoName 符合Go命名规范         eg: UserService
	RawServiceName string // 原始定义 - s.Desc.Name() proto中的原始定义 eg: user_service

	// ServiceTypeName 指定方法返回服务名，一般 si.PkgRefName + "." + si.ServiceName
	ServiceTypeName func() string

	Base            *ServiceInfo   // TODO 服务继承, 暂无场景(proto不支持，原thrift支持)
	CombineServices []*ServiceInfo // TODO 服务组合，暂无场景(proto不支持，原thrift支持)

	Methods      []*MethodInfo // 方法信息
	HasStreaming bool          // 是否包含流式方法，若Methods中包含流式方法，则为true

	// ServiceFilePath       string // 原kitex thrift使用，当前暂无场景
	Protocol              string // 原kitex thrift使用，当前暂无场景
	HandlerReturnKeepResp bool   // 原kitex thrift使用，当前暂无场景
	UseThriftReflection   bool   // 原kitex thrift使用，当前暂无场景

	// for multiple services scenario, the reference name for the service
	// RefName string // 服务引用名，若有同名service则import处应该要区分引用名. TODO 目前没设置，是否有该场景

	// identify whether this service would generate a corresponding handler.
	// 是否生成handler, proto中来源于 file.Generate, 如google.proto.api相关的IDL应该过滤掉
	// FIXME 此在初始过滤就应该过滤掉，如google.proto.api相关的IDL。 进入gen逻辑的应该全部为true
	GenerateHandler bool

	// ext by houzw
	Deprected bool       // 是否废弃
	Comment   CommentSet // 注释
}

// AllMethods returns all methods that the service have.
func (s *ServiceInfo) AllMethods() (ms []*MethodInfo) {
	ms = append(ms, s.Methods...)
	for base := s.Base; base != nil; base = base.Base { // 遍历链表
		ms = append(base.Methods, ms...)
	}
	return ms
}
