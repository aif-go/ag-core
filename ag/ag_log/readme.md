## ag_log
> 对slog的封装，ag-core使用slog作为统一的日志输出api，通过slog对三方日志框架进行插件式的封装，
> 目前已封装zap日志框架，后续会根据需要对其他日志框架进行封装。

### 设计思路
ag_log通过基础的slogmulti进行封装，slogmulti提供了fanout、failover、pool、route等模式的日志输出。
顶层的handler封装为fanout模式（golang计划原生支持fanout模式的Multihandler模式），再下层支持多种日志组合模式。
ag_log通过NamedHandler实现对handler的命名，通过名称实现对handler的多模式组织。


### FUTURE
- 动态日志级别支持
- 自定义日志服务支持