package agslog

import (
	"ag-core/ag/ag_conf"
	"fmt"
	"log/slog"
	"sync"

	slogmulti "github.com/samber/slog-multi"
)

const (
	AgSlogPropertiesKeyPrefix = "aglog"
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
	props *AgSlogProperties
	// handlers []slog.Handler // 直接注册的handler 好像没用
	// 直接注册的tophandler
	custTopHandlers []slog.Handler
	// handlerDefs   []HandlerDefinition
	// handlersCaches map[string]INamedHandler
	handlersCaches       sync.Map // 缓存每个原子handler
	handlers             sync.Map // 缓存每个等层handler，其pipline了中间件
	namedLogger          sync.Map
	replaceableHandllers sync.Map
	// namedHandlers  map[string]slog.Handler
	// 工厂
	factories []*HandlerFactory

	middlewares []slogmulti.Middleware

	logMu sync.Mutex // 用于保护namedLogger
}

// NewBuilder creates a new logger builder.
func NewBuilder() *Builder {
	return DefaultBuilder()
}

func newBuilder() *Builder {
	return &Builder{}
}

// WithProperties sets the properties for the logger.
func (b *Builder) WithProperties(props *AgSlogProperties) *Builder {
	b.props = props
	return b
}

// AddHandlerFactorys adds handler factories to the builder.
func (b *Builder) AddHandlerFactorys(factorys []*HandlerFactory) *Builder {
	// func (b *Builder) AddHandlerFactorys(factorys ...*HandlerFactory) *Builder {
	if len(factorys) == 0 {
		return b
	}
	b.AddHandlerFactory(factorys...)
	return b
}

// AddHandlerFactoryss adds handler factories to the builder.
func (b *Builder) AddHandlerFactoryss(factoryss [][]*HandlerFactory) *Builder {
	// func (b *Builder) AddHandlerFactoryss(factoryss ...[]*HandlerFactory) *Builder {
	if len(factoryss) == 0 {
		return b
	}
	for _, factorys := range factoryss {
		b.AddHandlerFactorys(factorys)
	}
	return b
}

// AddHandlerFactory adds handler factory to the builder.
func (b *Builder) AddHandlerFactory(factory ...*HandlerFactory) *Builder {
	if len(factory) == 0 {
		return b
	}
	for _, factory := range factory {
		b.addHandlerFactory(factory)
	}
	return b
}

func (b *Builder) addHandlerFactory(factory *HandlerFactory) *Builder {
	if factory == nil {
		return b
	}
	b.factories = append(b.factories, factory)
	return b
}

// AddHandlers adds handlers to the builder.
// func (b *Builder) AddHandlers(handlers ...slog.Handler) *Builder {
func (b *Builder) AddHandlers(handlers []slog.Handler) *Builder {
	if len(handlers) == 0 {
		return b
	}
	return b.AddHandler(handlers...)
}

// AddHandlerss adds handlers to the builder.
func (b *Builder) AddHandlerss(handlerss [][]slog.Handler) *Builder {
	if len(handlerss) == 0 {
		return b
	}
	for _, handlers := range handlerss {
		// err := b.AddHandlers(handlers...)
		b.AddHandlers(handlers)
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

	logger, err := b.initTopLogger()
	if err != nil {
		return nil, err
	}

	b.tryReplaceNamedHandler()

	return logger, nil
}

func (b *Builder) initTopLogger() (*slog.Logger, error) {
	// 解析顶层handler
	topHandlers, err := b.resolveTopHandlers()
	if err != nil {
		// fmt.Printf("agslog: resolve top handler fail: %v", err)
		return nil, err
	}

	// 打印顶级handler信息
	if len(topHandlers) == 0 {
		fmt.Println("agslog: no top handler specified, use default slog handler")
		// return TopLogger(), nil
		// topHandlers = append(topHandlers, TopLogger().Handler())
		// topHandlers = append(topHandlers, slog.Default().Handler())

		tlog := TopLogger()
		return tlog, nil
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
	var rhandler slog.Handler
	if len(topHandlers) > 1 {
		rhandler = slogmulti.Fanout(topHandlers...)
	} else {
		rhandler = topHandlers[0]
	}

	// 中间件（作用与顶层handler前）
	if len(b.middlewares) > 0 {
		rhandler = slogmulti.Pipe(b.middlewares...).Handler(rhandler)
		// FIXME
		// future: 动态日志级别支持
	}

	rhandler = b.wrapNamedHandlerIfNeed(topLoggerName, rhandler)

	// 替换top handler
	topLog := TopLogger()
	thandler := topLog.Handler()
	th, ok := thandler.(*ReplaceableHandler)
	if ok {
		// 检查handler是否符合name
		if !th.IsMatchesName() {
			th.ReplaceHandler(rhandler)
		}
	}

	// 是否设置当前log实现为slog默认实现，将直接替换slo的全局默认调用
	if b.props.IsDefault {
		slog.SetDefault(topLog)
	}

	return topLog, nil
}

func (b *Builder) tryReplaceNamedHandler() {
	b.replaceableHandllers.Range(func(k, v any) bool {
		if k == topLoggerName {
			return true // 不处理topLoggerName
		}

		// 只处理ReplaceableHandler
		rh, ok := v.(*ReplaceableHandler)
		if !ok {
			return true
		}

		// 检查handler是否符合name
		if rh.IsMatchesName() {
			return true
		}

		// 获取handler名称
		name := rh.Name()

		// 解析handler
		h, err := b.resolveHandler(name)
		if err != nil {
			th := TopLogger().Handler()
			// if rh.handler.Load() == th {
			if (*rh.handler.Load()) == th {
				return true
			}
			h = th // 若解析失败，则默认使用top handler
		}

		// 替换handler
		rh.ReplaceHandler(h)

		return true
	})

}

// AddHandler 添加handler
func (b *Builder) AddHandler(handler ...slog.Handler) *Builder {
	if len(handler) == 0 {
		return b
	}
	for _, h := range handler {
		name, err := b.addHandler(h)
		if err != nil {
			// slog.Error("AddHandler fail", "name", name, "err", err)
			slog.Error(fmt.Sprintf("AddHandler fail: %s err:%v", name, err))
		}
	}
	return b
}

// addHandler 添加handler
func (b *Builder) addHandler(handler slog.Handler) (name string, err error) {
	if nameh, ok := handler.(INamedHandler); ok {
		name = nameh.Name()
		// if _, exists := b.namedHandlers[name]; exists {
		// 	return name, fmt.Errorf("named handler [%s] already registered", name)
		// }
		// b.namedHandlers[name] = nameh
		if _, exists := b.handlersCaches.Load(name); exists {
			return name, fmt.Errorf("named handler [%s] already registered", name)
		}
		b.handlersCaches.Store(name, nameh)
		// b.handlers = append(b.handlers, nameh)

		fmt.Printf("agslog: regist handler name:[%s] type:[%T[%T]]\n", name, nameh, nameh.Original())
	} else {
		// b.handlers = append(b.handlers, handler)
		// 获取handler类型名
		name = fmt.Sprintf("%T", handler)
		// if _, exists := b.namedHandlers[name]; exists {
		// 	return name, fmt.Errorf("named handler [%s] already registered", name)
		// }
		// b.namedHandlers[name] = handler
		if _, exists := b.handlersCaches.Load(name); exists {
			return name, fmt.Errorf("named handler [%s] already registered", name)
		}
		b.handlersCaches.Store(name, handler)

		fmt.Printf("agslog: regist handler name:[%s] type:[%T]\n", name, handler)
	}
	return
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
			fmt.Printf("agslog[error]: resolveHandler[%s] fail:%v\n", name, err)
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
	// if handler, ok := b.namedHandlers[hname]; ok {
	// 		return handler, nil
	// }
	if handler, ok := b.handlersCaches.Load(hname); ok {
		return handler.(slog.Handler), nil
	}

	// 2. 工厂注册的handler
	for _, f := range b.factories {
		if f.Name == hname {
			handler, err := f.GetHandler(b.resolveHandler) // 递归调用
			if err != nil {
				return nil, err
			}

			b.handlersCaches.Store(hname, handler) // 缓存handler

			return handler, nil
		}
	}

	return nil, fmt.Errorf("handler %s not found", hname)
}

func (b *Builder) GetSlogByName(hname string) *slog.Logger {
	// 第一次检查
	if logger, ok := b.namedLogger.Load(hname); ok {
		return logger.(*slog.Logger)
	}

	// 加锁（每个名称独立的锁）
	b.logMu.Lock()
	defer b.logMu.Unlock()

	// 在锁内，第二次检查
	if logger, ok := b.namedLogger.Load(hname); ok {
		return logger.(*slog.Logger)
	}

	// handler, ok := b.handlers.Load(hname)
	// if ok {
	// 	return slog.New(handler.(slog.Handler))
	// }
	var handler slog.Handler
	handler = b.getNamedHandler(hname)
	if handler == nil {
		blogger := TopLogger()
		if hname == topLoggerName || blogger == nil {
			blogger = slog.Default() // 若是topLoggerName，则使用默认handler,否则TopLogger.Handler会循环调用自己
		}

		handler = blogger.Handler()
	}

	handler = NewReplaceableHandler(hname, handler)
	if hname != topLoggerName { // 不是topLoggerName，才缓存
		b.replaceableHandllers.Store(hname, handler)
	}

	logger := slog.New(handler)

	// 缓存logger
	b.namedLogger.Store(hname, logger)
	return logger
}

func (b *Builder) getNamedHandler(hname string) INamedHandler {
	handler, ok := b.handlers.Load(hname)
	if ok {
		return handler.(INamedHandler)
	}

	// 解析handler创建logger
	rh, err := b.resolveHandler(hname)
	if err != nil {
		slog.Warn(fmt.Sprintf("resolve log handler[%s] fail:%v", hname, err))
		return nil
	}
	if rh == nil {
		return nil
	}

	if len(b.middlewares) > 0 {
		rh = slogmulti.Pipe(b.middlewares...).Handler(rh)
	}
	handler = b.wrapNamedHandlerIfNeed(hname, rh)

	b.handlers.Store(hname, handler)
	return handler.(INamedHandler)
}

func (b *Builder) wrapNamedHandlerIfNeed(hname string, handler slog.Handler) slog.Handler {
	// if handler, ok := handler.(INamedHandler); ok {
	// 	return handler
	// }
	// return NewNamedHandler(hname, handler)
	return WrapNamedHandlerIfNeed(hname, handler)
}

// BuildAgSlog 创建slog logger
func BuildAgSlog(builder *Builder) (*slog.Logger, error) {
	logger, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func WrapNamedHandlerIfNeed(hname string, handler slog.Handler) slog.Handler {
	if handler, ok := handler.(INamedHandler); ok {
		if handler.Name() == hname {
			return handler
		}
		return NewNamedHandler(hname, handler)
	}
	return NewNamedHandler(hname, handler)
}
