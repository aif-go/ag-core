{{- if gt (len .CombineServices) 0}}
type {{call .ServiceTypeName}} interface {
{{- range .CombineServices}}
	{{.PkgRefName}}.{{.ServiceName}}
{{- end}}
}
{{- end}}

var {{LowerFirst .ServiceName}}Methods = map[string]serviceinfo.MethodInfo{
	{{- range .AllMethods}}
		"{{.RawName}}": serviceinfo.NewMethodInfo(
			{{LowerFirst .Name}}Handler,
			new{{.ArgStructName}},
			{{if .Oneway}}nil{{else}}new{{.ResStructName}}{{end}},
			{{if .Oneway}}true{{else}}false{{end}},
			serviceinfo.WithStreamingMode(
				{{- if and .ServerStreaming .ClientStreaming -}} serviceinfo.StreamingBidirectional
				{{- else if .ServerStreaming -}} serviceinfo.StreamingServer
				{{- else if .ClientStreaming -}} serviceinfo.StreamingClient
				{{- else -}}
					 {{- if or (eq $.Codec "protobuf") (eq .StreamingMode "unary") -}} serviceinfo.StreamingUnary
					 {{- else -}} serviceinfo.StreamingNone
					 {{- end -}}
				{{- end}}),
		),
	{{- end}}
}

func Register_{{.ServiceName}}_GRPCServer(srv {{.ServiceType}}Server) server.Option {
	return server.WithServiceRegistrar(&server.ServiceRegistrar{
		ServiceInfo: New{{.ServiceName}}ServiceInfo(),
		Handler: srv,
	})
}


// New{{.ServiceName}}ServiceInfo creates a new ServiceInfo containing all methods
{{- /* It's for the Server (providing both streaming/non-streaming APIs), or for the grpc client */}}
func New{{.ServiceName}}ServiceInfo() *serviceinfo.ServiceInfo {
	return new{{.ServiceName}}ServiceInfo({{- if .HasStreaming}}true{{else}}false{{end}}, true, true)
}

// New{{.ServiceName}}ServiceInfoForClient creates a new ServiceInfo containing non-streaming methods
{{- /* It's for the KitexThrift Client with only non-streaming APIs */}}
func New{{.ServiceName}}ServiceInfoForClient() *serviceinfo.ServiceInfo {
	return new{{.ServiceName}}ServiceInfo(false, false, true)
}

{{- /* It's for the StreamClient with only streaming APIs */}}
func New{{.ServiceName}}ServiceInfoForStreamClient() *serviceinfo.ServiceInfo {
	return new{{.ServiceName}}ServiceInfo(true, true, false)
}

func new{{.ServiceName}}ServiceInfo(hasStreaming bool, keepStreamingMethods bool, keepNonStreamingMethods bool) *serviceinfo.ServiceInfo {
	serviceName := "{{.RawServiceName}}"
	handlerType := (*{{call .ServiceTypeName}})(nil)
	methods := map[string]serviceinfo.MethodInfo{}
	for name, m := range {{LowerFirst .ServiceName}}Methods {
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
	svcInfo := &serviceinfo.ServiceInfo{
		ServiceName: 	 serviceName,
		HandlerType: 	 handlerType,
		Methods:     	 methods,
	{{- if ne "Hessian2" .ServiceInfo.Protocol}}
		PayloadCodec:  	 serviceinfo.{{.Codec | UpperFirst}},
	{{- end}}
		KiteXGenVersion: "{{.Version}}",
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
{{- if or (eq $.Codec "protobuf") ($isStreaming) }}
{{- $arg = index .Args 0}}{{/* streaming api only supports exactly one argument */}}
{{- end}}

func {{LowerFirst .Name}}Handler(ctx context.Context, handler interface{}, arg, result interface{}) error {
	{{- if eq $.Codec "protobuf"}} {{/* protobuf logic */}}
	{{- if $unary}} {{/* unary logic */}}
	switch s := arg.(type) {
	case *streaming.Args:
		st := s.Stream
		req := new({{NotPtr $arg.Type}})
		if err := st.RecvMsg(req); err != nil {
			return err
		}
		resp, err := handler.({{.ServiceName}}Server).{{.Name}}(ctx, req)
		if err != nil {
			return err
		}
		return st.SendMsg(resp)
	case *{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ArgStructName}}:
		success, err := handler.({{.ServiceName}}Server).{{.Name}}(ctx{{range .Args}}, s.{{.Name}}{{end}})
		if err != nil {
			return err
		}
		realResult := result.(*{{if not .GenArgResultStruct}}{{.PkgRefName}}.{{end}}{{.ResStructName}})
		realResult.Success = {{if .IsResponseNeedRedirect}}&{{end}}success
		return nil
	default:
		return errors.New("invalid message type for service method handler")
	}
	{{- else}}{{/* streaming logic */}}
		streamingArgs, ok := arg.(*streaming.Args)
		if !ok {
			return errors.New("invalid message type for service method handler")
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

{{- if and (eq $.Codec "protobuf") (not $.NoFastAPI)}} {{/* protobuf logic */}}
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

{{- if and (eq $.Codec "protobuf") (not $.NoFastAPI)}} {{/* protobuf logic */}}
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

type k{{.ServiceName}}Client struct {
	c client.Client
}

func new{{.ServiceName}}ServiceClient(c client.Client) *k{{.ServiceName}}Client {
	return &k{{.ServiceName}}Client{
		c: c,
	}
}

{{range .AllMethods}}
{{- if or .ClientStreaming .ServerStreaming}}
{{- /* streaming logic */}}
func (p *k{{.ServiceName}}Client) {{.Name}}(ctx context.Context{{if not .ClientStreaming}}{{range .Args}}, {{LowerFirst .Name}} {{.Type}}{{end}}{{end}}) ({{.ServiceName}}_{{.RawName}}Client, error) {
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
func (p *k{{.ServiceName}}Client) {{.Name}}(ctx context.Context {{range .Args}}, {{.RawName}} {{.Type}}{{end}}) ({{if not .Void}}r {{.Resp.Type}}, {{end}}err error) {
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

{{- if .FrugalPretouch}}
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

var Fx{{.ServiceName}}GRPCModule = fx.Module("fx_{{.ServiceName}}_GRPC",
	fx.Provide(
		fx.Annotate(
			Register_{{.ServiceName}}_GRPCServer,
			fx.ResultTags(`group:"ag_kitex_server_registrars"`),
		),
	),
)
