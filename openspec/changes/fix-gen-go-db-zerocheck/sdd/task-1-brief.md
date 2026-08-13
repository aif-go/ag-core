### [Task 1] Step 1: RED — 编写 getZeroCheck bool 类型测试
- 范围: req-001
- 设计来源:
  - specs/zerocheck-fix/spec.md#req-001
  - design.md#变更点
- 加载技能: devflow-golang-patterns, devflow-golang-commenting
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: 在 getZeroCheck 的 switch 中增加 `case "bool"`，返回 `!lowerStructName.jsonTag`
- Verify: `go build ./tool/cmd/gen-go-db/...`

### [Task 1] Step 2: GREEN — getZeroCheck 添加 bool 分支
- 范围: req-001
- 设计来源:
  - specs/zerocheck-fix/spec.md#req-001
  - design.md#变更点
- 加载技能: devflow-golang-patterns, devflow-golang-commenting
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: 在 getZeroCheck 的 switch 中增加 `case "bool"`，返回 `!lowerStructName.jsonTag`
- Verify: `go build ./tool/cmd/gen-go-db/...`

### [Task 2] Step 3: RED — 编写无普通列时 generalColZeroVal 测试
- 范围: req-002
- 设计来源:
  - specs/zerocheck-fix/spec.md#req-002
  - design.md#变更点
- 加载技能: devflow-golang-patterns, devflow-golang-commenting
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: generateListZeroValueColsMethod 中先遍历列判断是否存在普通列，有才声明 generalColZeroVal
- Verify: `go build ./tool/cmd/gen-go-db/...`

### [Task 2] Step 4: GREEN — 条件声明 generalColZeroVal
- 范围: req-002
- 设计来源:
  - specs/zerocheck-fix/spec.md#req-002
  - design.md#变更点
- 加载技能: devflow-golang-patterns, devflow-golang-commenting
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: generateListZeroValueColsMethod 中先遍历列判断是否存在普通列，有才声明 generalColZeroVal
- Verify: `go build ./tool/cmd/gen-go-db/...`

### [Task 5] Step 5: 验证 — 编译验证
- 范围: req-001, req-002
- 设计来源:
  - specs/zerocheck-fix/spec.md
  - design.md#变更点
- 文件: tool/cmd/gen-go-db/model/template.go
- 行为描述: 运行编译验证确认无错误
- Verify: `go build ./tool/cmd/gen-go-db/...`
