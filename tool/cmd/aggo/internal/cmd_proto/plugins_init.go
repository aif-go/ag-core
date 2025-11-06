package cmd_proto

const (
	PluginGO      = "go"
	PluginServer  = "server"
	PluginHertz   = "hertz"
	PluginKitex   = "kitex"
	PluginService = "service"
	PluginOpenapi = "openapi"
)

func init() {
	// base
	// RegPlugin(PluginBase, ModelBase, "--go_out=paths=source_relative:.")

	// go
	RegPlugin(PluginGO, ModelBase, "--go_out=paths=source_relative:.")

	// server
	RegPlugin(PluginServer, ModelServer, "--go-agserver_out=model=server:.")
	RegPlugin(PluginServer, ModelBase, "--go-agserver_out=model=xxx:.") // xxx实际不存在，这里只保持基础生成

	// service
	RegPlugin(PluginService, ModelServer, "--go-agservice_out=.")

	// kitex
	RegPlugin(PluginKitex, ModelServer, "--go-agkitex_out=model=server:.")
	RegPlugin(PluginKitex, ModelClient, "--go-agkitex_out=model=client:.")

	// hertz
	RegPlugin(PluginHertz, ModelServer, "--go-aghertz_out=model=server:.")
	RegPlugin(PluginHertz, ModelClient, "--go-aghertz_out=model=client:.")

	// openapi
	RegPlugin(PluginOpenapi, ModelBase, "--openapi_out=fq_schema_naming=true,default_response=false:.")
}
