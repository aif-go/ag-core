## 1. taskPool 借还封装

- [x] RED: 写失败测试 — async_worker_group_test.go → TestTaskPoolReturnResets、TestTaskPoolReturnNil
    Assertion: Return 后再 Borrow 的 task 的 ctx/handler 为 nil、record 为零值；Return(nil) 不 panic
    Expected failure: logTaskPool.Borrow/Return 未定义，编译失败
- [x] GREEN: 实现 logTaskPool 及 Borrow/Return — async_worker_group.go → (*logTaskPool).Borrow / (*logTaskPool).Return
    References RED test: TestTaskPoolReturnResets
    Verification: go test -run 'TestTaskPoolReturn' -count=1
- [x] REFACTOR: processTask 改走 taskPool.Return，Handle 的 taskPool.Get 改 Borrow — async_worker_group.go → processTask；async_handler.go → Handle

## 2. drop_new 丢弃归还

- [x] RED: 写失败测试 — async_worker_group_test.go → TestSubmitDropNewDropsAndReturns
    Assertion: 队列满时 Dropped+1、Queued 不变、task 归还到池（随后 Borrow 得到同一对象）
    Expected failure: 丢弃路径未归还 task，指针复用断言失败
- [ ] GREEN: drop_new 默认分支归还 task — async_worker_group.go → (*WorkerGroup).Submit
    References RED test: TestSubmitDropNewDropsAndReturns
    Verification: go test -run TestSubmitDropNewDropsAndReturns -count=1

## 3. drop_old 丢弃归还与非阻塞计数

- [x] RED: 写失败测试 — async_worker_group_test.go → TestSubmitDropOldReturnsOldQueuesNew、TestSubmitCounterReconciliation、TestSubmitUnknownStrategy
    Assertion: 队列满时 Dropped+1、旧 task 归还、新 task 入队并被处理；对账 Queued+Dropped==N 且 Processed==Queued；未知策略返回 nil 不 panic
    Expected failure: 旧 task 未归还、Queued 漏计数、Dropped 兜底漏计数
- [x] GREEN: drop_old 归还旧 task、入队改非阻塞 select、补 Queued/Dropped 计数 — async_worker_group.go → (*WorkerGroup).Submit
    References RED test: TestSubmitDropOldReturnsOldQueuesNew
    Verification: go test -run 'TestSubmitDropOld|TestSubmitCounterReconciliation|TestSubmitUnknownStrategy' -count=1

## 4. ctx 取消传播切断

- [x] RED: 写失败测试 — async_worker_group_test.go → TestHandleCutsCancellation、TestHandlePreservesCtxValues
    Assertion: 已取消 ctx 经 Handle 异步后 worker 收到 ctx 的 Done()==nil 且 Err()==nil；带 WithValue 的 ctx 异步后仍能读到值
    Expected failure: ctx 未做 WithoutCancel，worker 收到已取消的 ctx（Done 非 nil）
- [x] GREEN: Handle 使用 context.WithoutCancel — async_handler.go → (*AsyncHandler).Handle
    References RED test: TestHandleCutsCancellation
    Verification: go test -run 'TestHandleCutsCancellation|TestHandlePreservesCtxValues' -count=1

## 5. 验证

- [x] 5.1 go build ./...
- [x] 5.2 go test ./ag/ag_log/async/... -count=1
- [x] 5.3 go vet ./ag/ag_log/async/...
