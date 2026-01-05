package tpl

import (
	"ag-core/tool/aggen/generator"
	"ag-core/tool/aggen/types"
)

// ServerImportsSetter 设置Import部分信息
var ServerImportsSetter = func(geni *types.GennerInfo) error {
	// _module := geni.ModuleInfo
	// _pkg := geni.PkgInfo
	// _pkgI := geni.PackageInfo
	_svc := geni.ServiceInfo

	// FIXME 因为代码生成可能不是服务端应用，所以通过go_package获取import路径不准确，目前通过当前module名 + package名组合为import路径
	// if _module.HasPwdGoMod {
	// 	// 若当前存在go module则以当前gomodule路径为准
	// 	geni.AddImport(_pkg.PkgRefName, fmt.Sprintf("%s/%s", _module.PwdGoMod, _pkg.ImportPkg))
	// } else {
	// 	geni.AddImport(_pkg.PkgRefName, _pkg.ImportPath)
	// }

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

	geni.AddImports("context")
	geni.AddImport("app", "github.com/cloudwego/hertz/pkg/app")
	geni.AddImport("consts", "github.com/cloudwego/hertz/pkg/protocol/consts")
	// geni.AddImport("server", "ag-core/ag/ag_hertz/server")
	geni.AddImport("hserver", "ag-core/contribute/aghertz/server")

	return nil
}

// ServerTpl is the template for generating server.go.
var ServerTpl string = generator.Tpl_version +
	generator.Tpl_pkg +
	generator.Tpl_import +
	serverTpl +
	""

var serverTpl string = `
{{- $LGenI := .}}
{{- $LMI := .ModuleInfo }}     {{/* 模块信息 */}}
{{- $LGI := .GlobalInfo }}     {{/* 全局信息 */}}
{{- $LPkgInfo := .PkgInfo }}   {{/* 包信息 */}}
{{- $LPG := .PackageGroup }}   {{/* 包组信息 */}}
{{- $LPI := .PackageInfo }}    {{/* IDL文件级别信息 */}}
{{- $LSI := .ServiceInfo }}    {{/* 服务信息 */}}

{{- $LPkgRefName := $LPkgInfo.PkgRefName }} {{/*包名*/}}
{{- $LServiceName := $LSI.ServiceName }}{{/*服务名*/}}
{{- $LServiceTypeName := call $LSI.ServiceTypeName}} {{/*服务类型名(接口)*/}}

{{- range $LSI.AllMethods}}
	{{- $LMethod := .}} {{/*方法*/}}
	{{- $LArgs := index $LMethod.Args 0}} {{/* 入参目前规定只有一个 */}}
	{{- if not $LMethod.IsStreaming }} {{/* 非流方法才可以构建http服务 */}}
		{{- range $LMethod.HttpDescs}} {{/* 每个方法可能会有多个http描述 */}}

			{{- $LHttpDesc := .}}

			{{/*
			// Register_{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_{{.Method}}_Hertz
			// @Name: {{$LMethod.Name}}
			// @Method: {{.Method}}
			// @Path: {{.Path}}
			func Register_{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_{{.Method}}_Hertz(s {{$LServiceTypeName}}) server.Option {
				return server.WithRoute(&server.Route{
					HttpMethod:   "{{.Method}}",
					RelativePath: "{{.Path}}",
					Handlers:     append(make([]app.HandlerFunc, 0), _{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_HTTP_Handler(s)),
				})
			}
			*/}}
			// Router_{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_{{.Method}}_Hertz
			// @Name: {{$LMethod.Name}}
			// @Method: {{.Method}}
			// @Path: {{.Path}}
			func Router_{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_{{.Method}}_Hertz(s {{$LServiceTypeName}}) *hserver.Route {
				return &hserver.Route{
					HttpMethod:   "{{.Method}}",
					RelativePath: "{{.Path}}",
					Handlers:     append(make([]app.HandlerFunc, 0), _{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_HTTP_Handler(s)),
				}
			}

			func _{{$LServiceName}}_{{$LMethod.Name}}_{{.Num}}_HTTP_Handler(s {{$LServiceTypeName}}) func(ctx context.Context, c *app.RequestContext) {
				return func(ctx context.Context, c *app.RequestContext) {
					//var in = new({{$LArgs.UnptrType}})
					var in {{$LArgs.UnptrType}}

					{{- if $LHttpDesc.HasBody}}
						// BindByContentType 当Method为GET时，会自动调用BindQuery，其他方法会调用对应Bind
						if err := c.BindByContentType(&in{{$LHttpDesc.Body}}); err != nil {
						   c.String(consts.StatusBadRequest, err.Error())
							return
						}
					{{- else}}
						// 无Body定义的方法，BindQuery，否则使用BindByContentType
						if err := c.BindQuery(&in); err != nil {
						    c.String(consts.StatusBadRequest, err.Error())
							return
						}
					{{- end}}

					{{- if $LHttpDesc.HasVars}}
						if err := c.BindPath(&in); err != nil {
						    c.String(consts.StatusBadRequest, err.Error())
							return
						}
					{{- end}}

					resp, err := s.{{$LMethod.Name}}(ctx, &in)
					if err != nil {
						// TODO 增强异常处理
						c.String(consts.StatusInternalServerError, err.Error())
						return
					}

					// c.JSON 会自动将 Content-Type 设置为 application/json,
					c.JSON(consts.StatusOK, resp{{.ResponseBody}})
				}
			}

		{{- end}}
	{{- else }}
		// FIXME {{$LServiceName}}.{{$LMethod.Name}} is streaming method
	{{- end }}
{{- end }}
`
