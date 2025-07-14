{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}

{{- range .MethodSets}}
const Operation{{$svrType}}{{.OriginalName}} = "/{{$svrName}}/{{.OriginalName}}"
{{- end}}

{{- range .Methods}}
func Register_{{$svrType}}_{{.Name}}_HTTPServer(srv {{$svrType}}Server) server.Option {
	return server.WithRoute(&server.Route{
		HttpMethod:   "{{.Method}}",
		RelativePath: "{{.Path}}",
		Handlers:     append(make([]app.HandlerFunc, 0),_{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv)),
	})
}
{{- end}}

{{range .Methods}}
func _{{$svrType}}_{{.Name}}{{.Num}}_HTTP_Handler(srv {{$svrType}}Server) func(ctx context.Context, c *app.RequestContext) {
	return func(ctx context.Context, c *app.RequestContext) {
		var in = new({{.Request}})
		{{- if .HasBody}}
		if err := c.BindByContentType(in); err != nil {
		    c.String(consts.StatusBadRequest, err.Error())
			return
		}
		{{- end}}
		if err := c.BindQuery(in); err != nil {
		    c.String(consts.StatusBadRequest, err.Error())
			return
		}
		{{- if .HasVars}}
		if err := c.BindPath(in); err != nil {
		    c.String(consts.StatusBadRequest, err.Error())
			return
		}
		{{- end}}
		reply, err := srv.{{.Name}}(ctx, in)
		if err != nil {
			c.String(consts.StatusInternalServerError, err.Error())
			return
		}
		c.JSON(consts.StatusOK, reply{{.ResponseBody}})
	}
}
{{end}}

type {{.ServiceType}}HTTPClient interface {
{{- range .MethodSets}}
	{{.Name}}(ctx context.Context, req *{{.Request}}, opts ...config.RequestOption) (rsp *{{.Reply}}, err error)
{{- end}}
}

type {{.ServiceType}}HTTPClientImpl struct{
	cc *client.Client
}

func New{{.ServiceType}}HTTPClient (client *client.Client) {{.ServiceType}}HTTPClient {
	return &{{.ServiceType}}HTTPClientImpl{client}
}

{{range .MethodSets}}
func (c *{{$svrType}}HTTPClientImpl) {{.Name}}(ctx context.Context, in *{{.Request}}, opts ...config.RequestOption) (*{{.Reply}}, error) {
	var out {{.Reply}}
	path := "{{.Path}}"
	pathVars := make(map[string]string)
	{{- if .HasVars}}
	{{range .PathVars}}
    pathVars["{{.}}"] = in.Get{{.}}()
    {{- end}}
    {{- end}}
	{{if .HasBody -}}
	err := c.cc.Invoke(ctx, "{{.Method}}", path, pathVars, in{{.Body}}, &out{{.ResponseBody}}, opts...)
	{{else -}}
	err := c.cc.Invoke(ctx, "{{.Method}}", path, pathVars, nil, &out{{.ResponseBody}}, opts...)
	{{end -}}
	if err != nil {
		return nil, err
	}
	return &out, nil
}
{{end}}
var Fx{{.ServiceType}}HTTPModule = fx.Module("fx_{{.ServiceType}}_HTTP",
    fx.Provide(
        {{range .Methods}}
        fx.Annotate(
            Register_{{$svrType}}_{{.Name}}_HTTPServer,
			fx.ResultTags(`group:"hertz_router_options"`),
        ),
        {{end}}
    ),
)
