## 1. failover 核心组合逻辑

- [x] RED: 写失败测试 — slog_failover_test.go → TestFailoverFallsBackOnError、TestFailoverStopsOnFirstSuccess、TestFailoverAllFailReturnsError
    Assertion: getDoGetHandlerFunc 组合的 handler 按序切换（首选失败切次选且返回 nil / 首选成功不切 / 全失败返回 error）
    Expected failure: getDoGetHandlerFunc 未定义，编译失败
- [x] GREEN: 实现 failover 包核心 — slog_failover.go → getDoGetHandlerFunc（含 AgSlogFailoverProperties、BindAgSLogFailoverProperties、NewFailoverHandlerFactorys）
    References RED test: TestFailoverFallsBackOnError
    Verification: go test -run 'TestFailoverFallsBackOnError|TestFailoverStopsOnFirstSuccess|TestFailoverAllFailReturnsError' -count=1

## 2. failover 工厂 nil 安全

- [x] RED: 写失败测试 — slog_failover_test.go → TestNewFailoverHandlerFactorysNilSafe
    Assertion: nil props 与 nil Logs 均返回空 factory 列表且不 panic
    Expected failure: 工厂无 nil 保护，nil props 触发 panic
- [x] GREEN: 增加 nil 保护 — slog_failover.go → NewFailoverHandlerFactorys
    References RED test: TestNewFailoverHandlerFactorysNilSafe
    Verification: go test -run TestNewFailoverHandlerFactorysNilSafe -count=1

## 3. failover fx 装配与注册

- [x] 3.1 新增 zfx_aglog_failover.go — failover.FxAgSlogFailoverProvide（Bind + Annotate 到 group "agslog.factorys"）
- [x] 3.2 修改 zfx_aglog.go — FxAglogMode 注册 failover.FxAgSlogFailoverProvide

## 4. fanout nil 安全修复

- [x] RED: 写失败测试 — slog_fanout_test.go → TestBindAgSLogFanoutPropertiesBindError、TestNewFanoutHandlerFactorysNilSafe
    Assertion: Bind 失败返回非 nil 空结构且 error 为 nil；nil props/nil Logs 返回空列表不 panic
    Expected failure: Bind 失败返回 nil；工厂无 nil 保护 panic
- [x] GREEN: 修复 fanout — slog_fanout.go → BindAgSLogFanoutProperties、NewFanoutHandlerFactorys
    References RED test: TestBindAgSLogFanoutPropertiesBindError
    Verification: go test -run 'TestBindAgSLogFanoutPropertiesBindError|TestNewFanoutHandlerFactorysNilSafe' -count=1

## 5. 集成测试

- [x] 5.1 新增 test/agslog_failover.yaml — 含 aglog.failover.logs 配置（failover 名 → 有序子 handler 列表）及对应 zap 子 handler
- [x] 5.2 新增集成测试 test/agslog_failover_test.go → TestAgSlogFailover — fx 装配（含 failover.FxAgSlogFailoverProvide）+ GetSlogByName 解析 failover handler + 写日志验证无 error

## 6. 验证

- [x] 6.1 go build ./...
- [x] 6.2 go test ./ag/ag_log/... -count=1
- [x] 6.3 go vet ./ag/ag_log/...
