package tpl

import (
	"ag-core/tool/aggen/genkitex/tpl/kitextpl"
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
	"fmt"
)

// ServiceImportsSetter 设置Import部分信息
var ServiceImportsSetter = func(geni *types.GennerInfo) error {
	_module := geni.ModuleInfo
	_pkg := geni.PkgInfo
	_pkgI := geni.PackageInfo
	_svc := geni.ServiceInfo

	// util.TestLog("ServiceImportsSetter", _pkg)

	geni.AddImports("errors")
	geni.AddImport("client", "github.com/cloudwego/kitex/client")
	geni.AddImport("kitex", "github.com/cloudwego/kitex/pkg/serviceinfo")

	// pkg.AddImport(pkg.ServiceInfo.PkgRefName, pkg.ServiceInfo.ImportPath)
	if _module.HasPwdGoMod {
		// 若当前存在go module则以当前gomodule路径为准
		geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	} else {
		geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	}

	if len(_svc.AllMethods()) > 0 {
		geni.AddImports("context")
	}

	for _, m := range _svc.AllMethods() {
		// if !g.StreamX && (m.ClientStreaming || m.ServerStreaming) {
		if m.ClientStreaming || m.ServerStreaming {
			geni.AddImports("fmt")
		}
		if m.GenArgResultStruct {
			// if env.UsePrutalMarshal() {
			// 	// reuse "proto" so that no need to change code templates ...
			// 	pkg.AddImport("proto", "github.com/cloudwego/prutal")
			// } else {
			// 	pkg.AddImports("proto")
			// }
			geni.AddImport("proto", "google.golang.org/protobuf/proto") // TODO 暂不使用prutal，后续验证后考虑
		} else {
			// for method Arg and Result
			geni.AddImport(m.PkgRefName, m.ImportPath)
		}

		// TODO streaming 是否还是使用kitex的streaming
		// if m.ClientStreaming || m.ServerStreaming || (pkg.Codec == "protobuf" && !g.StreamX) {
		if m.ClientStreaming || m.ServerStreaming || _pkgI.Codec == "protobuf" {
			// protobuf handler support both PingPong and Unary (streaming) requests
			geni.AddImport("streaming", "github.com/cloudwego/kitex/pkg/streaming")
		}

		// 函数参数涉及到的依赖
		for _, a := range m.Args {
			for _, dep := range a.Deps {
				// 添加函数参数依赖
				if dep.PkgRefName != _pkg.PkgRefName { // 和上面import重复
					geni.AddImport(dep.PkgRefName, dep.ImportPath)
				}
			}
		}
		// 函数响应涉及到的依赖
		if !m.Void && m.Resp != nil {
			for _, dep := range m.Resp.Deps {
				if dep.PkgRefName != _pkg.PkgRefName { // 和上面import重复
					geni.AddImport(dep.PkgRefName, dep.ImportPath)
				}
			}
		}
		// 函数异常涉及到的依赖
		for _, e := range m.Exceptions {
			for _, dep := range e.Deps {
				if dep.PkgRefName != _pkg.PkgRefName { // 和上面import重复
					geni.AddImport(dep.PkgRefName, dep.ImportPath)
				}
			}
		}

	}

	return nil
}

// ServiceTpl is the template for generating the servicename.go source.
// 源：kitex/tool/internal_pkg/tpl/service.go ServiceTpl
var ServiceTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	serviceTpl +
	""

var serviceTpl string = `
{{ $LGenI := .}}
{{ $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{ $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{ $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{ $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{ $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{ $LSI := .ServiceInfo }}    {{/* 服务信息 */}}


{{- with $LSI }}
` + kitextpl.ServiceTpl + `
{{ end }}

`

/*
为尽量和kitex原模板保持一致，此处将serverInfo设置为当前处理对象，kitex此处的.实际上是PackageInfo，注意区别
当前模型下serviceInfo不再内嵌在PackageInfo下，下模板中原访问PackageInfo的部分调整为引用 $LPI 中获取

kitex原始模板: https://github.com/cloudwego/kitex/blob/v0.14.1/tool/internal_pkg/tpl/service.go#L18
*/

var serviceKitexTpl string = `
{{$HandlerReturnKeepResp := .HandlerReturnKeepResp}}
{{$UseThriftReflection := .UseThriftReflection}}

var errInvalidMessageType = errors.New("invalid message type for service method handler")

{{- if gt (len .CombineServices) 0}}
type {{call .ServiceTypeName}} interface {
{{- range .CombineServices}}
	{{.PkgRefName}}.{{.ServiceName}}
{{- end}}
}
{{- end}}

var {{LowerFirst .ServiceName}}ServiceMethods = map[string]kitex.MethodInfo{
	{{- range .AllMethods}}
		"{{.RawName}}": kitex.NewMethodInfo(
			{{LowerFirst .Name}}Handler,
			new{{.ArgStructName}},
			{{if .Oneway}}nil{{else}}new{{.ResStructName}}{{end}},
			{{if .Oneway}}true{{else}}false{{end}},
			kitex.WithStreamingMode(
				{{- if and .ServerStreaming .ClientStreaming -}} kitex.StreamingBidirectional
				{{- else if .ServerStreaming -}} kitex.StreamingServer
				{{- else if .ClientStreaming -}} kitex.StreamingClient
				{{- else -}}
					 {{- if or (eq $LPI.Codec "protobuf") (eq .StreamingMode "unary") -}} kitex.StreamingUnary
					 {{- else -}} kitex.StreamingNone
					 {{- end -}}
				{{- end}}),
		),
	{{- end}}
}

var (
	{{LowerFirst .ServiceName}}ServiceInfo = NewServiceInfo()
	{{LowerFirst .ServiceName}}ServiceInfoForClient = NewServiceInfoForClient()
	{{LowerFirst .ServiceName}}ServiceInfoForStreamClient = NewServiceInfoForStreamClient()
)

// for server
func serviceInfo() *kitex.ServiceInfo {
	return {{LowerFirst .ServiceName}}ServiceInfo
}

// for stream client
func serviceInfoForStreamClient() *kitex.ServiceInfo {
	return {{LowerFirst .ServiceName}}ServiceInfoForStreamClient
}

// for client
func serviceInfoForClient() *kitex.ServiceInfo {
	return {{LowerFirst .ServiceName}}ServiceInfoForClient
}

// NewServiceInfo creates a new ServiceInfo containing all methods
{{- /* It's for the Server (providing both streaming/non-streaming APIs), or for the grpc client */}}
func NewServiceInfo() *kitex.ServiceInfo {
	return newServiceInfo({{- if .HasStreaming}}true{{else}}false{{end}}, true, true)
}

// NewServiceInfo creates a new ServiceInfo containing non-streaming methods
{{- /* It's for the KitexThrift Client with only non-streaming APIs */}}
func NewServiceInfoForClient() *kitex.ServiceInfo {
	return newServiceInfo(false, false, true)
}

{{- /* It's for the StreamClient with only streaming APIs */}}
func NewServiceInfoForStreamClient() *kitex.ServiceInfo {
	return newServiceInfo(true, true, false)
}

func newServiceInfo(hasStreaming bool, keepStreamingMethods bool, keepNonStreamingMethods bool) *kitex.ServiceInfo {
	serviceName := "{{.RawServiceName}}"
	handlerType := (*{{call .ServiceTypeName}})(nil)
	methods := map[string]kitex.MethodInfo{}
	for name, m := range  {{LowerFirst .ServiceName}}ServiceMethods {
		if m.IsStreaming() && !keepStreamingMethods {
			continue
		}
		if !m.IsStreaming() && !keepNonStreamingMethods {
			continue
		}
		methods[name] = m
	}
	extra := map[string]interface{}{
		"PackageName":	 "{{.PkgInfo.PkgName}}",
		{{- if $UseThriftReflection }}
		"ServiceFilePath": {{ backquoted .ServiceFilePath }},
		{{end}}
	}
	{{- if gt (len .CombineServices) 0}}
	extra["combine_service"] = true
	extra["combined_service_list"] = []string{
		{{- range  .CombineServices}}"{{.ServiceName}}",{{- end}}
	}
	{{- end}}
	if hasStreaming {
		extra["streaming"] = hasStreaming
	}
	svcInfo := &kitex.ServiceInfo{
		ServiceName: 	 serviceName,
		HandlerType: 	 handlerType,
		Methods:     	 methods,
	{{- /*- if ne "Hessian2" .ServiceInfo.Protocol*/ -}}
	{{- if ne "Hessian2" .Protocol -}} {{/* FIXME by houzw */}}
		PayloadCodec:  	 kitex.{{$LPI.Codec | UpperFirst}},
	{{- end -}}
		{{/*KiteXGenVersion: "{{.Version}}",  TODO 如何获取kitex版本? by houzw */}}
		KiteXGenVersion: "v0.14.1_by_aggen",  // FIXME 模板生成时应该获取kitex版本
		Extra:           extra,
	}
	return svcInfo
}

{{range .AllMethods}}
{{- $isStreaming := or .ClientStreaming .ServerStreaming}}
{{- $unary := and (not .ServerStreaming) (not .ClientStreaming)}}
{{- $clientSide := and .ClientStreaming (not .ServerStreaming)}}
{{- $serverSide := and (not .ClientStreaming) .ServerStreaming}}
{{- $bidiSide := and .ClientStreaming .ServerStreaming}}
{{- $arg := "" }}
{{- if or (eq $LPI.Codec "protobuf") ($isStreaming) }}
{{- $arg = index .Args 0}}{{/* streaming api only supports exactly one argument */}}
{{- end}}

func {{LowerFirst .Name}}Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	{{- if eq $LPI.Codec "protobuf"}} {{/* protobuf logic */}}
	{{- if $unary}} {{/* unary logic */}}
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new({{NotPtr $arg.Type}})
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx, req)
		if err != nil {
			return err
		}
		return st.SendMsg(resp)
	case *{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ArgStructName}}:
		success, err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx{{range .Args}}, s.{{.Name}}{{end}})
		if err != nil {
			return err
		}
		realResult := result.(*{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}})
		realResult.Success = {{if .IsResponseNeedRedirect}}&{{end}}success
		return nil
	default:
		return errInvalidMessageType
	}
	{{- else}}{{/* streaming logic */}}
		streamingArgs, ok := arg.(*streaming.Args)
		if !ok {
			return errInvalidMessageType
		}
		st := streamingArgs.Stream
		stream := &{{LowerFirst .ServiceName}}{{.RawName}}Server{st}
		{{- if $serverSide}}
		req := new({{NotPtr $arg.Type}})
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		{{- end}}
		return handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}({{if $serverSide}}req, {{end}}stream)
	{{- end}} {{/* $unary end */}}
	{{- else}} {{/* thrift logic */}}
	{{- if $unary}} {{/* unary logic */}}
	{{- if eq .StreamingMode "unary"}}
	if streaming.GetStream(ctx) == nil {
		return errors.New("{{.ServiceName}}.{{.Name}} is a thrift streaming unary method, please call with Kitex StreamClient or remove the annotation streaming.mode")
	}
	{{- end}}
	{{if gt .ArgsLength 0}}realArg := {{else}}_ = {{end}}arg.(*{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ArgStructName}})
	{{if or (not .Void) .Exceptions}}realResult := result.(*{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}}){{end}}
	{{if .Void}}err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx{{range .Args}}, realArg.{{.Name}}{{end}})
	{{else}}success, err := handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(ctx{{range .Args}}, realArg.{{.Name}}{{end}})
	{{end -}}
	if err != nil {
	{{- if $HandlerReturnKeepResp }}
		// still keep resp when err is not nil
		// NOTE: use "-handler-return-keep-resp" to generate this
		{{if not .Void}}realResult.Success = {{if .IsResponseNeedRedirect}}&{{end}}success{{- end}}
	{{- end }}
	{{if .Exceptions -}}
		switch v := err.(type) {
		{{range .Exceptions -}}
		case {{.Type}}:
			 realResult.{{.Name}} = v
		{{end -}}
		default:
			 return err
		}
	} else {
	{{else -}}
		return err
	}
	{{end -}}
	{{if not .Void}}realResult.Success = {{if .IsResponseNeedRedirect}}&{{end}}success{{end}}
	{{- if .Exceptions}}}{{end}}
	return nil
	{{- else}} {{/* $isStreaming */}}
	st, ok := arg.(*streaming.Args)
	if !ok {
		return errors.New("{{.ServiceName}}.{{.Name}} is a thrift streaming method, please call with Kitex StreamClient")
	}
	stream := &{{LowerFirst .ServiceName}}{{.RawName}}Server{st.Stream}
		{{- if not $serverSide}}
	return handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(stream)
		{{- else}} {{/* !$serverSide */}}
		{{- $RequestType := $arg.Type}}
	req := new({{NotPtr $RequestType}})
	if err := st.Stream.RecvMsg(req); err != nil {
		return err
	}
	return handler.({{.PkgRefName}}.{{.ServiceName}}).{{.Name}}(req, stream)
		{{- end}} {{/* $serverSide end*/}}
    {{- end}} {{/* thrift end */}}
	{{- end}} {{/* protobuf end */}}
}

{{- /* define streaming struct */}}
{{- if $isStreaming}}
type {{LowerFirst .ServiceName}}{{.RawName}}Client struct {
	streaming.Stream
}

{{- /* For generating RPCFinish events when streaming.FinishStream(s, err) is called manually */}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) DoFinish(err error) {
	if finisher, ok := x.Stream.(streaming.WithDoFinish); ok {
		finisher.DoFinish(err)
	} else {
		panic(fmt.Sprintf("streaming.WithDoFinish is not implemented by %T", x.Stream))
	}
}

{{- if or $clientSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) Send(m {{$arg.Type}}) error {
	return x.Stream.SendMsg(m)
}
{{- end}}

{{- if $clientSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) CloseAndRecv() ({{.Resp.Type}}, error) {
	if err := x.Stream.Close(); err != nil {
		return nil, err
	}
	m := new({{NotPtr .Resp.Type}})
	return m, x.Stream.RecvMsg(m)
}
{{- end}}

{{- if or $serverSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Client) Recv() ({{.Resp.Type}}, error) {
	m := new({{NotPtr .Resp.Type}})
	return m, x.Stream.RecvMsg(m)
}
{{- end}}

type {{LowerFirst .ServiceName}}{{.RawName}}Server struct {
	streaming.Stream
}

{{if or $serverSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Server) Send(m {{.Resp.Type}}) error {
	return x.Stream.SendMsg(m)
}
{{end}}

{{if $clientSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Server) SendAndClose(m {{.Resp.Type}}) error {
	return x.Stream.SendMsg(m)
}
{{end}}

{{if or $clientSide $bidiSide}}
func (x *{{LowerFirst .ServiceName}}{{.RawName}}Server) Recv() ({{$arg.Type}}, error) {
	m := new({{NotPtr $arg.Type}})
	return m, x.Stream.RecvMsg(m)
}
{{end}}

{{- end}}
{{- /* define streaming struct end */}}
func new{{.ArgStructName}}() interface{} {
	return {{if not .GenArgResultStruct}}{{.PkgRefName}}.New{{.ArgStructName}}(){{else}}&{{.ArgStructName}}{}{{end}}
}
{{if not .Oneway}}
func new{{.ResStructName}}() interface{} {
	return {{if not .GenArgResultStruct}}{{.PkgRefName}}.New{{.ResStructName}}(){{else}}&{{.ResStructName}}{}{{end}}
}{{end}}

{{- if .GenArgResultStruct}}
{{$arg := index .Args 0}}
type {{.ArgStructName}} struct {
	Req {{$arg.Type}}
}

{{- if and (eq $LPI.Codec "protobuf") (not $LPI.NoFastAPI)}} {{/* protobuf logic */}}
func (p *{{.ArgStructName}}) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetReq() {
		p.Req = new({{NotPtr $arg.Type}})
	}
	return p.Req.FastRead(buf, _type, number)
}

func (p *{{.ArgStructName}}) FastWrite(buf []byte) (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.FastWrite(buf)
}

func (p *{{.ArgStructName}}) Size() (n int) {
	if !p.IsSetReq() {
		return 0
	}
	return p.Req.Size()
}
{{- end}} {{/* protobuf end */}}

func (p *{{.ArgStructName}}) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetReq() {
	   return out, nil
	}
	return proto.Marshal(p.Req)
}

func (p *{{.ArgStructName}}) Unmarshal(in []byte) error {
	msg := new({{NotPtr $arg.Type}})
	if err := proto.Unmarshal(in, msg); err != nil {
	   return err
	}
	p.Req = msg
	return nil
}

var {{.ArgStructName}}_Req_DEFAULT {{$arg.Type}}

func (p *{{.ArgStructName}}) GetReq() {{$arg.Type}} {
	if !p.IsSetReq() {
	   return {{.ArgStructName}}_Req_DEFAULT
	}
	return p.Req
}

func (p *{{.ArgStructName}}) IsSetReq() bool {
	return p.Req != nil
}

func (p *{{.ArgStructName}}) GetFirstArgument() interface{} {
	return p.Req
}

type {{.ResStructName}} struct {
	Success {{.Resp.Type}}
}

var {{.ResStructName}}_Success_DEFAULT {{.Resp.Type}}

{{- if and (eq $LPI.Codec "protobuf") (not $LPI.NoFastAPI)}} {{/* protobuf logic */}}
func (p *{{.ResStructName}}) FastRead(buf []byte, _type int8, number int32) (n int, err error) {
	if !p.IsSetSuccess() {
		p.Success = new({{NotPtr .Resp.Type}})
	}
	return p.Success.FastRead(buf, _type, number)
}

func (p *{{.ResStructName}}) FastWrite(buf []byte) (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.FastWrite(buf)
}

func (p *{{.ResStructName}}) Size() (n int) {
	if !p.IsSetSuccess() {
		return 0
	}
	return p.Success.Size()
}
{{- end}} {{/* protobuf end */}}

func (p *{{.ResStructName}}) Marshal(out []byte) ([]byte, error) {
	if !p.IsSetSuccess() {
	   return out, nil
	}
	return proto.Marshal(p.Success)
}

func (p *{{.ResStructName}}) Unmarshal(in []byte) error {
	msg := new({{NotPtr .Resp.Type}})
	if err := proto.Unmarshal(in, msg); err != nil {
	   return err
	}
	p.Success = msg
	return nil
}

func (p *{{.ResStructName}}) GetSuccess() {{.Resp.Type}} {
	if !p.IsSetSuccess() {
	   return {{.ResStructName}}_Success_DEFAULT
	}
	return p.Success
}

func (p *{{.ResStructName}}) SetSuccess(x interface{}) {
	p.Success = x.({{.Resp.Type}})
}

func (p *{{.ResStructName}}) IsSetSuccess() bool {
	return p.Success != nil
}

func (p *{{.ResStructName}}) GetResult() interface{} {
	return p.Success
}
{{- end}}
{{end}}

type kClient struct {
	c client.Client
}

func newServiceClient(c client.Client) *kClient {
	return &kClient{
		c: c,
	}
}

{{range .AllMethods}}
{{- if or .ClientStreaming .ServerStreaming}}
{{- /* streaming logic */}}
func (p *kClient) {{.Name}}(ctx context.Context{{if not .ClientStreaming}}{{range .Args}}, {{LowerFirst .Name}} {{.Type}}{{end}}{{end}}) ({{.ServiceName}}_{{.RawName}}Client, error) {
	streamClient, ok := p.c.(client.Streaming)
	if !ok {
		return nil, fmt.Errorf("client not support streaming")
	}
	res := new(streaming.Result)
	err := streamClient.Stream(ctx, "{{.RawName}}", nil, res)
	if err != nil {
		return nil, err
	}
	stream := &{{LowerFirst .ServiceName}}{{.RawName}}Client{res.Stream}
	{{if not .ClientStreaming -}}
	{{$arg := index .Args 0}}
	if err := stream.Stream.SendMsg({{LowerFirst $arg.Name}}); err != nil {
		return nil, err
	}
	if err := stream.Stream.Close(); err != nil {
		return nil, err
	}
	{{end -}}
	return stream, nil
}
{{- else}}
func (p *kClient) {{.Name}}(ctx context.Context {{range .Args}}, {{.RawName}} {{.Type}}{{end}}) ({{if not .Void}}r {{.Resp.Type}}, {{end}}err error) {
	var _args {{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ArgStructName}}
	{{range .Args -}}
	_args.{{.Name}} = {{.RawName}}
	{{end -}}
{{if .Void -}}
	{{if .Oneway -}}
	if err = p.c.Call(ctx, "{{.RawName}}", &_args, nil); err != nil {
		return
	}
	{{else -}}
	var _result {{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}}
	if err = p.c.Call(ctx, "{{.RawName}}", &_args, &_result); err != nil {
		return
	}
	{{if .Exceptions -}}
	switch {
	{{range .Exceptions -}}
	case _result.{{.Name}} != nil:
		return _result.{{.Name}}
	{{end -}}
	}
	{{end -}}
	{{end -}}
	return nil
{{else -}}
	var _result {{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}}
	if err = p.c.Call(ctx, "{{.RawName}}", &_args, &_result); err != nil {
		return
	}
	{{if .Exceptions -}}
	switch {
	{{range .Exceptions -}}
	case _result.{{.Name}} != nil:
		return r, _result.{{.Name}}
	{{end -}}
	}
	{{end -}}
	return _result.GetSuccess(), nil
{{end -}}
}
{{- end}}
{{end}}

{{- if $LPI.FrugalPretouch}}
var pretouchOnce sync.Once
func pretouch() {
	pretouchOnce.Do(func() {
{{- if gt (len .AllMethods) 0}}
		var err error
{{- range .AllMethods}}
		err = frugal.Pretouch(reflect.TypeOf({{if not .GenArgResultStruct}}{{.PkgRefName}}.New{{.ArgStructName}}(){{else}}&{{.ArgStructName}}{}{{end}}))
		if err != nil {
			goto PRETOUCH_ERR
		}
{{- if not .Oneway}}
		err = frugal.Pretouch(reflect.TypeOf({{if not .GenArgResultStruct}}{{.PkgRefName}}.New{{.ResStructName}}(){{else}}&{{.ResStructName}}{}{{end}}))
		if err != nil {
			goto PRETOUCH_ERR
		}
{{- end}}
{{- end}}{{/* range .AllMethods */}}
	return
PRETOUCH_ERR:
	println("Frugal pretouch in {{.ServiceName}} failed: " + err.Error())
{{- end}}{{/* if gt (len .AllMethods) 0 */}}
	})
}
{{- end}}{{/* if .FrugalPretouch */}}
`
