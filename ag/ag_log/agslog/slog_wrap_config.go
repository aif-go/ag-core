package agslog

import (
	"ag-core/ag/ag_conf"
	"fmt"
	"log/slog"
	"os"

	slogmulti "github.com/samber/slog-multi"
)

const (
	AgSlogPropertiesKeyPrefix = "aglog"
)

var (
	// AgSlog will be set to the created logger if IsDefault is true in properties.
	AgSlog = slog.Default()
)

// AgSlogProperties agslog配置
type AgSlogProperties struct {
	// 是否设置为默认log，默认为true
	IsDefault bool `value:"${:true}"`
	// 顶层handler名称
	TopHandler []string
}

// BindAgSlogProperties 绑定构建AgSlogProperties配置
func BindAgSlogProperties(binder ag_conf.IBinder) (*AgSlogProperties, error) {
	prop := &AgSlogProperties{}
	err := binder.Bind(prop, AgSlogPropertiesKeyPrefix)
	if err != nil {
		return nil, err
	}
	return prop, nil
}

// Builder for creating a slog.Logger.
type Builder struct {
	props    *AgSlogProperties
	handlers []slog.Handler
	// 直接注册的tophandler
	custTopHandlers []slog.Handler
	// handlerDefs   []HandlerDefinition
	namedHandlers map[string]INamedHandler
	// 工厂
	factories []*HandlerFactory

	middlewares []slogmulti.Middleware
}

// NewBuilder creates a new logger builder.
func NewBuilder() *Builder {
	return &Builder{
		namedHandlers: make(map[string]INamedHandler),
	}
}

// WithProperties sets the properties for the logger.
func (b *Builder) WithProperties(props *AgSlogProperties) *Builder {
	b.props = props
	return b
}

// AddHandlerFactorys adds handler factories to the builder.
func (b *Builder) AddHandlerFactorys(factorys ...*HandlerFactory) *Builder {
	if len(factorys) == 0 {
		return b
	}
	b.factories = append(b.factories, factorys...)
	return b
}

// AddHandlers adds handlers to the builder.
func (b *Builder) AddHandlers(handlers ...slog.Handler) *Builder {
	if len(handlers) == 0 {
		return b
	}

	for _, handler := range handlers {
		err := b.addHandler(handler)
		if err != nil {
			fmt.Printf("agslog: handler add fail: %v", err)
		}
	}
	return b
}

// AddMiddlewares adds middlewares to the builder.
func (b *Builder) AddMiddlewares(middlewares ...slogmulti.Middleware) *Builder {
	b.middlewares = append(b.middlewares, middlewares...)
	return b
}

// Deprecated
// RegTopHandler 注册顶层命名handler
// 不应该直接注册top handler
func (b *Builder) RegTopHandler(handler ...slog.Handler) {
	if len(handler) == 0 {
		return
	}
	b.custTopHandlers = append(b.custTopHandlers, handler...)
}

// Build creates the slog.Logger.
func (b *Builder) Build() (*slog.Logger, error) {
	if b.props == nil {
		b.props = &AgSlogProperties{IsDefault: true} // Default properties
	}

	// 解析顶层handler
	topHandlers, err := b.resolveTopHandlers()
	if err != nil {
		// fmt.Printf("agslog: resolve top handler fail: %v", err)
		return nil, err
	}

	// 打印顶级handler信息
	if len(topHandlers) == 0 {
		fmt.Println("agslog: no top handler specified, use json handler with level info to stdout")
		topHandlers = append(topHandlers, slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		fmt.Printf("agslog: top handler specified, use %d handler(s)\n", len(topHandlers))
		for _, handler := range topHandlers {
			name := "unknown"
			if nameh, ok := handler.(INamedHandler); ok {
				name = nameh.Name()
				fmt.Printf("agslog: top handler name:[%s] type:[%T[%T]]\n", name, nameh, nameh.Original())
			} else {
				fmt.Printf("agslog: top handler name:[%s] type:[%T]\n", name, handler)
			}
		}
	}

	// 顶层的handler是以fanout的方式组合的
	// FIXME slog将来将支持mutiHandler模式，且为fanout模式，届时考虑使用原生方式替换
	rhandler := slogmulti.Fanout(topHandlers...)

	// 中间件（作用与顶层handler前）
	if len(b.middlewares) > 0 {
		rhandler = slogmulti.Pipe(b.middlewares...).Handler(rhandler)
	}

	logger := slog.New(rhandler)

	// 是否设置当前log实现为slog默认实现，将直接替换slo的全局默认调用
	if b.props.IsDefault {
		slog.SetDefault(logger)
		AgSlog = logger
	}

	return logger, nil
}

// addHandler 添加handler
func (b *Builder) addHandler(handler slog.Handler) error {
	if nameh, ok := handler.(INamedHandler); ok {
		name := nameh.Name()
		if _, exists := b.namedHandlers[name]; exists {
			return fmt.Errorf("named handler [%s] already registered", name)
		}
		b.namedHandlers[name] = nameh
		b.handlers = append(b.handlers, nameh)
		fmt.Printf("agslog: registered handler name:[%s] type:[%T[%T]]\n", name, nameh, nameh.Original())
	} else {
		b.handlers = append(b.handlers, handler)
		fmt.Printf("agslog: registered handler name:[unknown] type:[%T]\n", handler)
	}
	return nil
}

// resolveTopHandlers 解析顶层handler
// 此方法应该在builder设置完全设置值后调用，TODO 添加强制控制，避免非法使用
func (b *Builder) resolveTopHandlers() ([]slog.Handler, error) {
	var topHandlers []slog.Handler

	if len(b.custTopHandlers) > 0 {
		topHandlers = append(topHandlers, b.custTopHandlers...)
	}

	if b.props == nil || len(b.props.TopHandler) == 0 {
		return topHandlers, nil
	}

	for _, name := range b.props.TopHandler {
		handler, err := b.resolveHandler(name) // 解析指定名称的handler
		if err != nil {
			fmt.Printf("agslog: resolveHandler[%s] fail:%v\n", name, err)
			continue
			// return nil, err
		}
		topHandlers = append(topHandlers, handler)
		// fmt.Printf("agslog: top handler %s not found, skip\n", name)
	}

	return topHandlers, nil
}

// resolveHandler 根据提供的handler的名称，获取对应handler
func (b *Builder) resolveHandler(hname string) (slog.Handler, error) {
	// 1. 直接注册的handler
	if handler, ok := b.namedHandlers[hname]; ok {
		return handler, nil
	}

	// 2. 工厂注册的handler
	for _, f := range b.factories {
		if f.Name == hname {
			handler, err := f.GetHandler(b.resolveHandler) // 递归调用
			if err != nil {
				return nil, err
			}
			return handler, nil
		}
	}

	return nil, fmt.Errorf("handler %s not found", hname)
}

// BuildAgSlog 创建slog logger
func BuildAgSlog(builder *Builder) (*slog.Logger, error) {
	logger, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
