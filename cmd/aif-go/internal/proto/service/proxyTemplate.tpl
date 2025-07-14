{{$ifcName := .Name}}
package service

import (
	"context"
	"log"
	"time"

	pb "{{.PbPkg}}"
	mw "github.com/frochyzhang/ag-core/ag/ag_ext"
)

// ===================== 代理实现 =====================
type {{$ifcName}}Proxy struct {
	service *{{$ifcName}}Service // 原始服务实例
	handlers map[string]mw.HandlerFunc
}

func New{{$ifcName}}Proxy(service *{{$ifcName}}Service, mws []mw.PrioritizedMiddleware) pb.{{$ifcName}}Server {
	proxy := &{{$ifcName}}Proxy{
		service:     service,
		handlers:    make(map[string]mw.HandlerFunc),
	}
    {{range .Methods}}
	proxy.handlers["{{.Name}}"] = mw.RegisterHandler("{{.Name}}", mws, func(ctx context.Context, req interface{}) (interface{}, error) {
		// 最终调用原始服务方法
		return proxy.service.{{.Name}}(ctx, req.(*pb.{{.Request}}))
	})
    {{- end}}
	return proxy
}

// ======== {{.Name}} 代理方法 ========{{range .Methods}}
func (p *{{$ifcName}}Proxy) {{.Name}}(ctx context.Context, in *pb.{{.Request}}) (*pb.{{.Reply}}, error) {
    start := time.Now()
    methodName := "{{.Name}}"

    // 获取处理链
	handler := p.handlers[methodName]

    // 执行调用链
    res, err := handler(ctx, in)
    if err != nil {
        log.Printf("[%s] failed in %v: %v", methodName, time.Since(start), err)
        return nil, err
    }

    log.Printf("[%s] success in %v", methodName, time.Since(start))
    return res.(*pb.{{.Reply}}), nil
}{{end}}