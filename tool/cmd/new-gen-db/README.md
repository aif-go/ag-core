# new-gen-db 使用说明书

## 概述

`new-gen-db` 是一个数据库相关文件生成工具，支持从 Excel 表格生成 YAML 配置文件，以及从 YAML 文件生成 Model 和 DAO 代码。

## 安装

```bash
cd tool/cmd/new-gen-db
go build -o gendb main.go
```

## 命令结构

```
gendb
├── yaml    # 从 Excel 生成 YAML 文件
└── db      # 从 YAML 生成 Model 和 DAO 文件
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

## 完整工作流程

### 典型使用流程

1. **准备 Excel 文件**：创建包含表结构定义的 Excel 文件

2. **生成 YAML 配置**：
   ```bash
   gendb yaml -i ./tables.xlsx -o ./config/yaml
   ```

3. **生成数据库代码**：
   ```bash
   gendb db -i ./config/yaml -o ./internal/database
   ```

### 快速测试流程

```bash
# 1. 生成示例 YAML 文件
gendb yaml --test -o ./test_yaml

# 2. 从示例 YAML 生成代码
gendb db -i ./test_yaml -o ./test_output
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
```

### Q: Excel 文件的格式要求是什么？

Excel 文件需要按照特定的格式定义表结构，包括字段名、类型、约束等信息。具体格式请参考项目中的 `test.xlsx` 示例文件。

### Q: 生成的代码支持哪些数据库？

目前支持 MySQL 和 DB2 两种数据库。可以通过 `--dbtype` 参数指定生成哪种数据库的代码，不指定则同时生成两种。

### Q: 如何自定义模块名称？

有两种方式：
1. 使用 `--module` 参数显式指定
2. 在包含 `go.mod` 文件的目录下运行，工具会自动读取模块名称

## 技术支持

如有问题或建议，请联系项目维护者。
