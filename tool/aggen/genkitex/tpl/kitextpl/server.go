// Copyright 2022 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Modifications made by ag-cmd houzw in 2025.09:
// - kitex原始模板: https://github.com/cloudwego/kitex/blob/v0.14.1/tool/internal_pkg/tpl/server.go
// - 移除import和package相关生成逻辑，由新方式生成
// - 生成内容方与kitex原模板一致，调整适配aggen的数据结构，原kitex传入对象为PackageInfo中组合了ServiceInfo，aggen不是
// - 注释掉了 模板扩展点 FIXME 后续有扩展需求再说

package kitextpl

// ServerTpl is the template for generating server.go.
var ServerTpl string = `

// NewServer creates a server.Server with the given handler and options.
func NewServer(handler {{call .ServiceTypeName}}, opts ...server.Option) server.Server {
    var options []server.Option
    {{- /* {{template "@server.go-NewServer-option" .}} */}}
    options = append(options, opts...)
	{{- if or (eq $LPI.Codec "thrift") $LPI.StreamX}}
	options = append(options, server.WithCompatibleMiddlewareForUnary())
	{{- end}}

    svr := server.NewServer(options...)
    if err := svr.RegisterService(serviceInfo(), handler); err != nil {
            panic(err)
    }
	{{- if $LPI.FrugalPretouch}}
	pretouch()
	{{- end}}{{/* if .FrugalPretouch */}}
    return svr
}
{{- /* {{template "@server.go-EOF" .}} */}}

func RegisterService(svr server.Server, handler {{call .ServiceTypeName}}, opts ...server.RegisterOption) error {
	return svr.RegisterService(serviceInfo(), handler, opts...)
}
`
