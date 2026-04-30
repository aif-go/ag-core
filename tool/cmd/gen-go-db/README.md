# gen-go-db 使用说明书

## 概述

`gen-go-db` 是一个数据库相关文件生成工具，支持从 Excel 表格生成 YAML 配置文件，从 YAML 文件生成 Model 和 DAO 代码，以及按关键字拆分 Excel Sheet。

## 安装

```bash
cd tool/cmd/gen-go-db
go build -o gendb main.go
```

## 命令结构

```
gendb
├── yaml    # 从 Excel 生成 YAML 文件
├── db      # 从 YAML 生成 Model 和 DAO 文件
└── sheet   # 按关键字拆分 Excel Sheet
```

## 子命令说明

### 1. yaml 子命令

从 Excel 电子表格生成 YAML 配置文件。

#### 参数

| 参数 | 简写 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `--input` | `-i` | string | 是* | - | 输入 Excel 文件的路径 |
| `--output` | `-o` | string | 是* | - | 存放生成的 YAML 文件的目录 |
| `--test` | `-t` | bool | 否 | false | 测试模式，生成示例 YAML 文件 |
| `--table` | `-T` | string | 否 | - | 指定表名，只生成该表的文件 |

> *注：在测试模式下，`--input` 和 `--output` 可以为空

#### 使用示例

**基本用法：从 Excel 生成 YAML**

```bash
gendb yaml --input ./data.xlsx --output ./yaml_output
```

**简写形式：**

```bash
gendb yaml -i ./data.xlsx -o ./yaml_output
```

**只生成指定表的 YAML：**

```bash
gendb yaml -i ./data.xlsx -o ./yaml_output -T user_table
```

**测试模式：生成示例 YAML 文件**

```bash
gendb yaml --test --output ./sample_yaml
```

### 2. db 子命令

从 YAML 配置文件生成 Model 和 DAO 代码文件。

#### 参数

| 参数 | 简写 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `--input` | `-i` | string | 是 | - | 输入 YAML 文件或目录的路径 |
| `--output` | `-o` | string | 是 | - | 存放生成的 Model 和 DAO 文件的目录 |
| `--table` | `-T` | string | 否 | - | 指定表名，只生成该表的文件 |
| `--module` | `-m` | string | 否 | go.mod | 模块名称，未指定则查找当前位置的 go.mod |
| `--dbtype` | `-d` | string | 否 | 两者都生成 | 数据库类型，可选值：`mysql`、`db2` |

#### 使用示例

**基本用法：从 YAML 生成 Model 和 DAO**

```bash
gendb db --input ./yaml_output --output ./generated
```

**简写形式：**

```bash
gendb db -i ./yaml_output -o ./generated
```

**只生成指定表的代码：**

```bash
gendb db -i ./yaml_output -o ./generated -T user_table
```

**指定模块名称：**

```bash
gendb db -i ./yaml_output -o ./generated -m myproject
```

**只生成 MySQL 相关代码：**

```bash
gendb db -i ./yaml_output -o ./generated -d mysql
```

**只生成 DB2 相关代码：**

```bash
gendb db -i ./yaml_output -o ./generated -d db2
```

### 3. sheet 子命令

按关键字拆分 Excel Sheet，将一个 Sheet 拆分为两部分：DDL 部分和自定义规则部分。

#### 参数

| 参数 | 简写 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|------|--------|------|
| `--input` | `-i` | string | 是 | - | 输入 Excel 文件的路径 |
| `--output` | `-o` | string | 否 | 原文件名_日期 | 输出拆分后 Excel 文件的路径或目录 |
| `--keyword` | `-k` | string | 是 | - | 拆分 Sheet 的关键字 |

#### 功能说明

- **关键字行处理**：包含关键字的行将被丢弃，不会出现在任何新 Sheet 中
- **Sheet 命名规则**：
  - 关键字之前的 Sheet：保持原 Sheet 名称
  - 关键字之后的 Sheet：原名称 + `_custom_rule` 后缀
- **格式规则**：
  - 所有单元格使用宋体，字号 11
  - 如果某行有任何值，则该行所有单元格都加黑色边框（即使中间的列是空白）
  - 如果某行的单元格内容包含汉字，则该行所有内容加粗
  - 列宽自动调整：根据单元格内容长度自动调整，汉字算 2 个字符宽度，最小宽度 10，最大宽度 40
- **输出文件名**：
  - 如果 `--output` 是目录或没有扩展名，则使用默认文件名：`原文件名_日期`（日期格式：20060102）
  - 如果指定了完整文件路径，则使用指定的文件名

#### 使用示例

**基本用法：拆分 Excel Sheet**

```bash
gendb sheet --input ./data.xlsx --output ./output.xlsx --keyword "自定义脚本名字"
```

**简写形式：**

```bash
gendb sheet -i ./data.xlsx -o ./output.xlsx -k "自定义脚本名字"
```

**输出到目录（自动生成文件名）：**

```bash
gendb sheet -i ./data.xlsx -o ./output_dir -k "自定义脚本名字"
# 生成文件：./output_dir/data_20260320.xlsx
```

## 完整工作流程

### 典型使用流程

#### 方案一：生成数据库代码流程

1. **准备 Excel 文件**：创建包含表结构定义的 Excel 文件

2. **生成 YAML 配置**：
   ```bash
   gendb yaml -i ./tables.xlsx -o ./config/yaml
   ```

3. **生成数据库代码**：
   ```bash
   gendb db -i ./config/yaml -o ./internal/database
   ```

#### 方案二：拆分 Excel Sheet 流程

1. **准备 Excel 文件**：创建包含表结构和自定义规则的 Excel 文件

2. **按关键字拆分 Sheet**：
   ```bash
   gendb sheet -i ./tables.xlsx -o ./output -k "自定义脚本名字"
   ```

3. **使用拆分后的文件**：
   - DDL 部分：用于生成 YAML 配置
   - 自定义规则部分：用于其他业务逻辑

### 快速测试流程

```bash
# 1. 生成示例 YAML 文件
gendb yaml --test -o ./test_yaml

# 2. 从示例 YAML 生成代码
gendb db -i ./test_yaml -o ./test_output
```

### 综合使用示例

```bash
# 1. 从原始 Excel 拆分出 DDL 和自定义规则部分
gendb sheet -i ./original.xlsx -o ./ddl_only.xlsx -k "自定义脚本名字"

# 2. 从 DDL 部分生成 YAML 配置
gendb yaml -i ./ddl_only.xlsx -o ./config/yaml

# 3. 从 YAML 生成数据库代码
gendb db -i ./config/yaml -o ./internal/database
```

## 常见问题

### Q: 如何查看帮助信息？

```bash
# 查看主命令帮助
gendb --help

# 查看 yaml 子命令帮助
gendb yaml --help

# 查看 db 子命令帮助
gendb db --help

# 查看 sheet 子命令帮助
gendb sheet --help
```

### Q: Excel 文件的格式要求是什么？

Excel 文件需要按照特定的格式定义表结构，包括字段名、类型、约束等信息。具体格式请参考项目中的 `test.xlsx` 示例文件。

### Q: 生成的代码支持哪些数据库？

目前支持 MySQL 和 DB2 两种数据库。可以通过 `--dbtype` 参数指定生成哪种数据库的代码，不指定则同时生成两种。

### Q: 如何自定义模块名称？

有两种方式：
1. 使用 `--module` 参数显式指定
2. 在包含 `go.mod` 文件的目录下运行，工具会自动读取模块名称

### Q: sheet 命令的关键字行是如何处理的？

包含关键字的行将被丢弃，不会出现在任何新 Sheet 中。关键字之前的行会放在第一个 Sheet（保持原名称），关键字之后的行会放在第二个 Sheet（名称为原名称 + `_custom_rule`）。

### Q: sheet 命令生成的 Excel 文件格式是什么样的？

生成的 Excel 文件具有以下格式特性：
- 所有单元格使用宋体，字号 11
- 如果某行有任何值，则该行所有单元格都加黑色边框（即使中间的列是空白）
- 如果某行的单元格内容包含汉字，则该行所有内容加粗
- 列宽自动调整：根据单元格内容长度自动调整，汉字算 2 个字符宽度，最小宽度 10，最大宽度 40

### Q: sheet 命令的输出文件名如何生成？

如果 `--output` 参数是目录或没有扩展名，则使用默认文件名：`原文件名_日期`（日期格式：20060102）。如果指定了完整文件路径，则使用指定的文件名。

## 技术支持

如有问题或建议，请联系项目维护者。
