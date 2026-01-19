package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ServiceImportsSetter 设置Import部分信息
var ProxyImportsSetter_v2 = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	// _pkg := geni.PkgInfo
	// _pkgI := geni.PackageInfo
	_svc := geni.ServiceInfo

	if !_module.HasPwdGoMod {
		return fmt.Errorf("pwdGoMod is empty")
	}

	for _, m := range _svc.Methods {
		// 入参
		for _, arg := range m.Args {
			for _, dep := range arg.Deps {
				geni.AddImport(dep.PkgRefName, dep.ImportPath)
			}
		}
		// 出参
		resp := m.Resp
		for _, dep := range resp.Deps {
			geni.AddImport(dep.PkgRefName, dep.ImportPath)
		}
	}

	// api 接口包
	geni.AddImport(_svc.PkgRefName, _svc.ImportPath)

	// geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))

	geni.AddImports("context")

	geni.AddImport("fmt", "fmt")

	serviceRef := fmt.Sprintf("%s/internal/service", _module.PwdGoMod)
	geni.AddImport("service", serviceRef) // 模块servcie import
	// geni.AddImport("smw", "ag-core/ag/ag_ext") // 服务中间件扩展
	// geni.AddImport("smw", "ag-core/ag/ag_service")
	geni.AddImport("ag_service", "ag-core/ag/ag_service")
	geni.AddImport("slog", "log/slog")

	return nil
}

// ServiceTpl is the template for generating the servicename.go source.
var ProxyTpl_v2 string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	proxyTpl_v2 +
	""

var proxyTpl_v2 string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- with $LSI }}
` + proxyInnerTpl_v2 + `
{{- end }}
`

var proxyInnerTpl_v2 string = `
{{- $LPkgRefName := $LPkgInfo.PkgRefName}}
{{- $LServiceInfo := $LSI}}
{{- $LServiceName := $LServiceInfo.ServiceName}}
{{- $LServiceImplName := printf "%sImpl" $LServiceInfo.ServiceName}}
{{- $LServiceProxyName := printf "%sProxy" $LServiceInfo.ServiceName}}

// 服务信息，接口级别
var {{$LServiceName}}ServiceInfo = ag_service.ServiceInfo{
	PackageName: "{{$LPkgInfo.PkgName}}",
	ServiceName: "{{$LServiceName}}",
	HandlerType: (*{{$LPkgRefName}}.{{$LServiceInfo.ServiceName}})(nil),
}

{{- range $LServiceInfo.AllMethods}}
	{{- $LMethod := .}}

    var {{$LServiceName}}{{$LMethod.Name}}CallInfo = ag_service.CallInfo{
        ServiceInfo: {{$LServiceName}}ServiceInfo,
	    CallName:    "{{$LMethod.Name}}",
        ClientStreaming: {{$LMethod.ClientStreaming}},
        ServerStreaming: {{$LMethod.ServerStreaming}},
	    Extra: map[string]interface{}{
	    	"method_name": "{{$LMethod.Name}}",
	    },
    }
{{- end}}

var {{$LServiceName}}CallInfos = map[string]ag_service.CallInfo{
    {{- range $LServiceInfo.AllMethods}}
    	{{- $LMethod := .}}
        "{{$LMethod.Name}}": {{$LServiceName}}{{$LMethod.Name}}CallInfo,
    {{- end}}
}

// {{$LServiceProxyName}} 服务代理
type {{$LServiceProxyName}} struct {
    ag_service.AgServiceProxyBase
	impl *service.{{$LServiceImplName}}
}

// New{{$LServiceProxyName}} 创建{{$LServiceProxyName}}
func New{{$LServiceProxyName}}(
    impl *service.{{$LServiceImplName}},
    agServiceBuilder ag_service.AgServiceBuilder,
    mws []ag_service.MiddlewareProvider,
  ) ({{$LPkgRefName}}.{{$LServiceInfo.ServiceName}}, error) {

    p :=  &{{$LServiceProxyName}}{
		impl:     impl,
        AgServiceProxyBase: ag_service.NewAgServiceProxyBase(
			{{$LServiceName}}ServiceInfo,
			{{$LServiceName}}CallInfos,
		),
	}

    for _, cif := range p.CallInfos {
    	if cif.ClientStreaming || cif.ServerStreaming {
			slog.Warn("agServiceProxy buildskip streaming method", "call_name", cif.CallName)
			continue
		}

		// 根据方法信息构建Endpoints
		oriEndpoint, err := p.getOriginalHandler(cif)
        if err != nil {
			return nil, err
		}
		endpoint, err := agServiceBuilder.BuildEndpointChain(
			cif,
			mws,
            oriEndpoint,
		)
		if err != nil {
			return nil, err
		}
		// 注册endpoint
		p.RegisterEndpoint(cif.CallName, endpoint)
	}
		
	return p, nil
}

{{- range $LServiceInfo.AllMethods}}
	{{- $LMethod := .}}

	{{- $LArg := index $LMethod.Args 0 }} {{/* 规定接口参数只有一个参数 */}}
	{{- $LResp := $LMethod.Resp}}
	{{- if and $LMethod.ClientStreaming $LMethod.ServerStreaming }}
		// {{$LMethod.Name}} {{$LServiceProxyName}} proxy ClientStreaming ServerStreaming
		func (p *{{$LServiceProxyName}}) {{$LMethod.Name}}(stream {{$LPkgRefName}}.{{$LServiceName}}_{{$LMethod.Name}}Server) error{
			// FIXME 流式接口暂不支持代理，直接调用实现
			return p.impl.{{$LMethod.Name}}(stream)
		}
	{{- else if $LMethod.ClientStreaming}}
		// {{$LMethod.Name}} {{$LServiceProxyName}} proxy ClientStreaming
		func (p *{{$LServiceProxyName}}) {{$LMethod.Name}}(stream {{$LPkgRefName}}.{{$LServiceName}}_{{$LMethod.Name}}Server) error{
			// FIXME 流式接口暂不支持代理，直接调用实现
			return p.impl.{{$LMethod.Name}}(stream)
		}
	{{- else if $LMethod.ServerStreaming}}
		// {{$LMethod.Name}} {{$LServiceProxyName}} proxy ServerStreaming
		func (p *{{$LServiceProxyName}}) {{$LMethod.Name}}(req {{$LArg.Type}}, stream {{$LPkgRefName}}.{{$LServiceName}}_{{$LMethod.Name}}Server) error{
			// FIXME 流式接口暂不支持代理，直接调用实现
			return p.impl.{{$LMethod.Name}}(req, stream)
		}
	{{- else}}
		// {{$LMethod.Name}} {{$LServiceProxyName}} proxy
		func (p *{{$LServiceProxyName}}) {{$LMethod.Name}}(ctx context.Context, req {{$LArg.Type}}) ({{$LResp.Type}}, error){
            endpoint := p.GetEndpoint("{{$LMethod.Name}}")

			resp,err := endpoint(ctx,req)
			if err!= nil {
				slog.Error("failed to handle request!", "method", "{{$LMethod.Name}}", "error", err)
				return nil, err
			}

			return resp.({{$LResp.Type}}),nil
		}
	{{- end}}
{{- end}}


// getOriginalHandler
func (p *{{$LServiceProxyName}}) getOriginalHandler(cif ag_service.CallInfo) (ag_service.Endpoint, error) {
    switch cif.CallName {
    {{- range $LServiceInfo.AllMethods}}
	    {{- $LMethod := .}}
        {{- $LArg := index $LMethod.Args 0 }} {{/* 规定接口参数只有一个参数 */}}
	    {{- $LResp := $LMethod.Resp}}
        case "{{$LMethod.Name}}":
	    {{- if and $LMethod.ClientStreaming $LMethod.ServerStreaming }}
	    	// return p.impl.{{$LMethod.Name}}(stream), nil
            return nil, fmt.Errorf("can not get original handler for streaming method: %s", cif.CallName)
	    {{- else if $LMethod.ClientStreaming}}
	    	// return p.impl.{{$LMethod.Name}}(stream), nil
            return nil, fmt.Errorf("can not get original handler for streaming method: %s", cif.CallName)
	    {{- else if $LMethod.ServerStreaming}}
	    	// FIXME 流式接口暂不支持代理，不允许获取原始接口
	    	// return p.impl.{{$LMethod.Name}}(req, stream), nil
            return nil, fmt.Errorf("can not get original handler for streaming method: %s", cif.CallName)
	    {{- else}}
		    return func(ctx context.Context, req interface{}) (interface{}, error) {
		    	return p.impl.{{$LMethod.Name}}(ctx, req.({{$LArg.Type}}))
		    }, nil
	    {{- end}}
    {{- end}}
    default:
		return nil, fmt.Errorf("handler for method %s not implemented", cif.CallName)
	}
}

`
