package service

import ({{ range .Packages }}
	"{{.}}"
	{{- end }}
	"go.uber.org/fx"
	mw "github.com/frochyzhang/ag-core/ag/ag_ext"
)

var FxServiceModule = fx.Module("fx-service",
	fx.Provide({{ range .Names }}
		New{{ .Name }}Service,
		New{{ .Name }}ProxyWithParams,
	{{- end }}
	),
)

type BaseFxMiddlewareParams struct {
	fx.In
	GlobalMws []mw.PrioritizedMiddleware              `group:"fx_global_service_middleware" ,optional:"true"`
}
{{ range .Names }}
type Fx{{.Name}}Middleware struct {
	BaseFxMiddlewareParams
	CustomMws []mw.PrioritizedMiddleware              `group:"fx_{{toLower .Name}}_service_middleware" ,optional:"true"`
}

func New{{.Name}}ProxyWithParams(in Fx{{.Name}}Middleware,service *{{.Name}}Service) {{ .Brief }}.{{ .Name }}Server {
	 return New{{.Name}}Proxy(service,append(in.GlobalMws,in.CustomMws...))
}
{{- end}}