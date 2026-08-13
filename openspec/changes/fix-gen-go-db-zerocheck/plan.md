# Execution Plan

## Dispatch Guardrails
### 粒度检查
- Task 计数 = 3 → 通过

### 执行纪律
- 每完成一次 dispatch → 执行 review（行为正确 + 构建通过）
- review 通过前 → 不得 dispatch 下一个

### 完成检查（全部任务完成后逐项验证）
- [ ] `go build ./tool/cmd/gen-go-db/...` 通过

---

## Steps

### [Task 1] Step 1: 修复 — getZeroCheck 添加 bool 分支
- 范围: req-001
- 设计来源:
  - specs/zerocheck-fix/spec.md#req-001
  - design.md#变更点
- 加载技能: devflow-golang-patterns, devflow-golang-commenting
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: 在 getZeroCheck 的 switch 中增加 `case "bool"`，返回 `!lowerStructName.jsonTag`
- Verify: `go build ./tool/cmd/gen-go-db/...`

### [Task 2] Step 2: 修复 — 条件声明 generalColZeroVal
- 范围: req-002
- 设计来源:
  - specs/zerocheck-fix/spec.md#req-002
  - design.md#变更点
- 加载技能: devflow-golang-patterns, devflow-golang-commenting
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: generateListZeroValueColsMethod 中先遍历列判断是否存在普通列，有才声明 generalColZeroVal
- Verify: `go build ./tool/cmd/gen-go-db/...`

### [Task 3] Step 3: 验证 — 编译验证
- 范围: req-001, req-002
- 设计来源:
  - specs/zerocheck-fix/spec.md
  - design.md#变更点
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: 运行编译验证确认无错误
- Verify: `go build ./tool/cmd/gen-go-db/...`

---

## Execution Mode Selection
REQUIRED: Use devflow-sdd skill.
DO NOT use executing-plans or inline execution.
