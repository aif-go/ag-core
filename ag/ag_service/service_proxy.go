package ag_service

type ServiceProxyBase struct {
	ServiceInfo ServiceInfo
	Endpoints   map[string]Endpoint
	MethodInfos map[string]CallInfo
}
