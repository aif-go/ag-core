## 偏差记录

### 2026-07-28 — 前置准备
- **偏差类型**: 跳过 worktree 创建和 baseline commit
- **原因**: 本次变更为 ag-core 框架工具代码修改（非服务项目），已在主仓库直接操作。仓库在执行前已有 9 个 dirty 路径，无法创建干净 baseline。
- **纠正**: 所有修改仅涉及 tool/cmd/gen-go-db/model/template.go 一个文件，无代码冲突风险。通过 `git diff` 管理变更。
- **是否影响结果**: 否
