package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ServiceImportsSetter 设置Import部分信息
var ProxyImportsSetter = func(geni *types.GennerInfo) error {
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

	serviceRef := fmt.Sprintf("%s/internal/service", _module.PwdGoMod)
	geni.AddImport("service", serviceRef) // 模块servcie import
	// geni.AddImport("smw", "ag-core/ag/ag_ext") // 服务中间件扩展
	geni.AddImport("smw", "ag-core/ag/ag_service")
	geni.AddImport("slog", "log/slog")

	return nil
}

// ServiceTpl is the template for generating the servicename.go source.
var ProxyTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	proxyTpl +
	""

var proxyTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- with $LSI }}
` + proxyInnerTpl + `
{{- end }}
`

var proxyInnerTpl string = `
{{- $LPkgRefName := $LPkgInfo.PkgRefName}}
{{- $LServiceInfo := $LSI}}
{{- $LServiceName := $LServiceInfo.ServiceName}}
{{- $LServiceImplName := printf "%sImpl" $LServiceInfo.ServiceName}}
{{- $LServiceProxyName := printf "%sProxy" $LServiceInfo.ServiceName}}

// {{$LServiceProxyName}} 服务代理
type {{$LServiceProxyName}} struct {
	impl *service.{{$LServiceImplName}}
	handlers map[string]smw.HandlerFunc
}

// New{{$LServiceProxyName}} 创建{{$LServiceProxyName}}
func New{{$LServiceProxyName}}(service *service.{{$LServiceImplName}}, mws []smw.PrioritizedMiddleware) {{$LPkgRefName}}.{{$LServiceInfo.ServiceName}} {
	proxy := &{{$LServiceProxyName}}{
		impl:     service,
		handlers: make(map[string]smw.HandlerFunc),
	}
	
	// 注册处理链
	{{- range $LServiceInfo.AllMethods}}
		{{- $LMethod := .}}
		{{- $LArg := index $LMethod.Args 0 }} {{/* 规定接口参数只有一个参数 */}}
		{{- $LResp := $LMethod.Resp}}

		{{- if and $LMethod.ClientStreaming $LMethod.ServerStreaming }}
			// FIXME {{$LMethod.Name}} 流式接口暂不支持代理
		{{- else if $LMethod.ClientStreaming}}
			// FIXME {{$LMethod.Name}} 流式接口暂不支持代理
		{{- else if $LMethod.ServerStreaming}}
			// FIXME {{$LMethod.Name}} 流式接口暂不支持代理
		{{- else}}
			proxy.handlers["{{$LMethod.Name}}"] = smw.RegisterHandler("{{$LMethod.Name}}", mws, func(ctx context.Context, req interface{}) (interface{}, error) {
				return proxy.impl.{{$LMethod.Name}}(ctx, req.({{$LArg.Type}}))
			})
		{{- end}}

	{{- end}}
		
	return proxy
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
			methodName := "{{$LMethod.Name}}"
			handler := p.handlers[methodName]

			res,err := handler(ctx,req)
			if err!= nil {
				slog.Error("failed to handle request!", "method", methodName, "error", err)
				return nil,err
			}

			return res.({{$LResp.Type}}),nil
		}
	{{- end}}

{{- end}}

`
