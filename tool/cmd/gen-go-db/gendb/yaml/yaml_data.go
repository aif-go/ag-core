package yaml

import (
	"github.com/iancoleman/orderedmap"
	"gopkg.in/yaml.v3"
)

type YamlAllTableData struct {
	YamlDataList []*YamlDataConfig
}

// 顶级配置结构（SelfQueryRules改为yaml.Node，用于有序解析）
type YamlDataConfig struct {
	DatabaseTable  DatabaseTable `yaml:"database_table"`
	SelfQueryRules yaml.Node     `yaml:"self_query_rules"` // 用yaml.Node保留顺序
}

// MarshalYAML 自定义YAML序列化方法，使用yaml.Node确保字段顺序
func (c YamlDataConfig) MarshalYAML() (interface{}, error) {
	node := yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}

	// 添加 database_table 字段
	databaseTableKey := yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "database_table"}
	databaseTableValue, err := c.DatabaseTable.MarshalYAML()
	if err != nil {
		return nil, err
	}
	node.Content = append(node.Content, &databaseTableKey, databaseTableValue.(*yaml.Node))

	// 添加 self_query_rules 字段
	selfQueryRulesKey := yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "self_query_rules"}
	node.Content = append(node.Content, &selfQueryRulesKey, &c.SelfQueryRules)

	return &node, nil
}

// 数据库表结构（包含表名、列、主键）
type DatabaseTable struct {
	TableName    string                 `yaml:"table_name"`
	TableComment string                 `yaml:"table_comment,omitempty"`
	Columns      *orderedmap.OrderedMap `yaml:"columns"`
	PrimaryKeys  []PrimaryKey           `yaml:"primary_keys"`
	Indexes      Indexes                `yaml:"indexes"`
	DbType       string                 `yaml:"-"`
}

// UnmarshalYAML implements custom YAML unmarshaling for DatabaseTable
func (dt *DatabaseTable) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// 创建一个临时结构体来处理解析，避免递归
	type tempDatabaseTable struct {
		TableName    string       `yaml:"table_name"`
		TableComment string       `yaml:"table_comment,omitempty"`
		Columns      yaml.Node    `yaml:"columns"`
		PrimaryKeys  []PrimaryKey `yaml:"primary_keys"`
		Indexes      Indexes      `yaml:"indexes"`
	}

	var temp tempDatabaseTable
	if err := unmarshal(&temp); err != nil {
		return err
	}

	// 赋值基本字段
	dt.TableName = temp.TableName
	dt.TableComment = temp.TableComment
	dt.PrimaryKeys = temp.PrimaryKeys
	dt.Indexes = temp.Indexes

	// 手动解析Columns字段以保持顺序
	dt.Columns = orderedmap.New()

	// 确保Columns节点是映射类型
	if temp.Columns.Kind == yaml.MappingNode {
		// 遍历YAML节点的内容（键值对）
		for i := 0; i < len(temp.Columns.Content); i += 2 {
			if i+1 < len(temp.Columns.Content) {
				keyNode := temp.Columns.Content[i]
				valueNode := temp.Columns.Content[i+1]

				// 获取键名
				key := keyNode.Value

				// 解析值为Column结构
				var column Column
				valueBytes, err := yaml.Marshal(valueNode)
				if err != nil {
					return err
				}

				if err := yaml.Unmarshal(valueBytes, &column); err != nil {
					return err
				}

				// 添加到orderedmap
				dt.Columns.Set(key, column)
			}
		}
	}

	return nil
}

// MarshalYAML implements custom YAML marshaling for DatabaseTable
// 使用orderedmap自动保持列的顺序
func (dt DatabaseTable) MarshalYAML() (interface{}, error) {
	node := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}

	// 添加table_name字段
	tableNameNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "table_name",
	}
	node.Content = append(node.Content, tableNameNode)

	tableNameValueNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: dt.TableName,
	}
	node.Content = append(node.Content, tableNameValueNode)

	// 添加table_comment字段（如果存在）
	if dt.TableComment != "" {
		tableCommentNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "table_comment",
		}
		node.Content = append(node.Content, tableCommentNode)

		tableCommentValueNode := &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: dt.TableComment,
		}
		node.Content = append(node.Content, tableCommentValueNode)
	}

	// 添加columns字段，orderedmap会自动保持插入顺序
	columnsNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "columns",
	}
	node.Content = append(node.Content, columnsNode)

	// 创建列的映射节点，保持orderedmap的顺序
	columnsMappingNode := &yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}

	// 遍历orderedmap中的键值对，保持顺序
	for _, key := range dt.Columns.Keys() {
		if value, ok := dt.Columns.Get(key); ok {
			// 添加键节点
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: key,
			}
			columnsMappingNode.Content = append(columnsMappingNode.Content, keyNode)

			// 添加值节点
			valueBytes, err := yaml.Marshal(value)
			if err != nil {
				return nil, err
			}

			var valueNode yaml.Node
			if err := yaml.Unmarshal(valueBytes, &valueNode); err != nil {
				return nil, err
			}

			columnsMappingNode.Content = append(columnsMappingNode.Content, valueNode.Content[0])
		}
	}

	node.Content = append(node.Content, columnsMappingNode)

	// 添加primary_keys字段
	primaryKeysNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "primary_keys",
	}
	node.Content = append(node.Content, primaryKeysNode)

	primaryKeysValueNode, err := yaml.Marshal(dt.PrimaryKeys)
	if err != nil {
		return nil, err
	}

	var primaryKeysNodeValue yaml.Node
	if err := yaml.Unmarshal(primaryKeysValueNode, &primaryKeysNodeValue); err != nil {
		return nil, err
	}

	node.Content = append(node.Content, primaryKeysNodeValue.Content[0])

	// 添加indexes字段
	indexesNode := &yaml.Node{
		Kind:  yaml.ScalarNode,
		Tag:   "!!str",
		Value: "indexes",
	}
	node.Content = append(node.Content, indexesNode)

	indexesValueNode, err := yaml.Marshal(dt.Indexes)
	if err != nil {
		return nil, err
	}

	var indexesNodeValue yaml.Node
	if err := yaml.Unmarshal(indexesValueNode, &indexesNodeValue); err != nil {
		return nil, err
	}

	node.Content = append(node.Content, indexesNodeValue.Content[0])

	return node, nil
}

// 列定义
type Column struct {
	DbColumn      string `yaml:"db_column"`
	GoType        string `yaml:"go_type"`
	Comment       string `yaml:"comment,omitempty"`
	PrimaryKey    bool   `yaml:"primary_key,omitempty"`
	NotNull       bool   `yaml:"not_null,omitempty"`
	Length        string `yaml:"length,omitempty"`
	DefaultValue  string `yaml:"default_value,omitempty"`
	Description   string `yaml:"description,omitempty"`
	AutoIncrement bool   `yaml:"auto_increment,omitempty"`
}

// 主键
type PrimaryKey struct {
	Column string `yaml:"column"`
}

// 单个索引（扁平化改造：BindColumns → Columns）
type Index struct {
	IndexName string   `yaml:"index_name"`
	Columns   []string `yaml:"columns"`
}

// 索引（普通+唯一）- 无需修改
type Indexes struct {
	General []Index `yaml:"general,omitempty"`
	Unique  []Index `yaml:"unique,omitempty"`
}

// 自定义查询规则
type QueryRule struct {
	SelectFields string       `yaml:"select_fields"`
	Aggregation  *Aggregation `yaml:"aggregation,omitempty"`
	OrderBy      string       `yaml:"order_by,omitempty"`
	Where        *WhereNode   `yaml:"where,omitempty"`
	Page         bool         `yaml:"page,omitempty"`
	DbType 	     string       `yaml:"dbtype,omitempty"`
}

// 聚合函数配置
type Aggregation struct {
	Function   string `yaml:"function"`
	ResultType string `yaml:"result_type"`
}

// WHERE条件节点（递归结构）
type WhereNode struct {
	Operator   string       `yaml:"operator,omitempty"`
	Conditions *[]WhereNode `yaml:"conditions,omitempty"`
	Expr       string       `yaml:"expr,omitempty"`
}

// MarshalYAML 自定义YAML序列化方法，使用yaml.Node确保字段顺序
func (w WhereNode) MarshalYAML() (interface{}, error) {
	// 创建yaml.Node来保持字段顺序
	node := yaml.Node{
		Kind: yaml.MappingNode,
		Tag:  "!!map",
	}

	// 按照特定顺序添加字段
	// 1. expr字段（如果存在）
	if w.Expr != "" {
		exprKey := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "expr",
		}
		exprValue := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: w.Expr,
		}
		node.Content = append(node.Content, &exprKey, &exprValue)
	}

	// 2. operator字段（如果存在）
	if w.Operator != "" {
		operatorKey := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "operator",
		}
		operatorValue := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: w.Operator,
		}
		node.Content = append(node.Content, &operatorKey, &operatorValue)
	}

	// 3. conditions字段（如果存在）
	if w.Conditions != nil && len(*w.Conditions) > 0 {
		conditionsKey := yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: "conditions",
		}

		// 序列化conditions数组
		conditionsValue, err := yaml.Marshal(w.Conditions)
		if err != nil {
			return nil, err
		}

		var conditionsNode yaml.Node
		if err := yaml.Unmarshal(conditionsValue, &conditionsNode); err != nil {
			return nil, err
		}

		node.Content = append(node.Content, &conditionsKey, conditionsNode.Content[0])
	}

	return &node, nil
}

// OrderedQueryRule 有序的查询规则（方法名+配置）
type OrderedQueryRule struct {
	MethodName string
	Rule       QueryRule
}

// NamingSqlData 自定义SQL数据
type NamingSqlData struct {
	MethodName       string          `yaml:"method_name"`
	NamingSql        string          `yaml:"naming_sql"`
	DbType           string          `yaml:"db_type,omitempty"`
	ParamColNameList []SqlParameter  `yaml:"param_col_name_list,omitempty"`
	SelectColumns    []*SelectColumn `yaml:"select_columns,omitempty"`
}

// SqlParameter SQL参数
type SqlParameter struct {
	ColName       string `yaml:"col_name"`
	ParameterName string `yaml:"parameter_name"`
	IsSlice       bool   `yaml:"is_slice"`
}

// SelectColumn 查询列
type SelectColumn struct {
	ColumnName string `yaml:"column_name"`
	Alias      string `yaml:"alias"`
}

// GenericColumn represents a generic column definition from an Excel sheet
type GenericColumn struct {
	Name          string `json:"name"`
	Type          string `json:"type"`
	Length        string `json:"length"`
	NotNull       string `json:"not_null"`
	DefaultValue  string `json:"default_value"`
	AutoIncrement string `json:"auto_increment"`
	Description   string `json:"description"`
	Tag           string `json:"custom_type"`
	IsPrimaryKey  bool   `json:"is_primary_key"` // 标记是否为主键
}

// GenericRule represents a generic query rule from an Excel sheet
type GenericRule struct {
	Name            string     `json:"name"`
	SelectFields    string     `json:"select_fields"`
	Conditions      *WhereNode `json:"conditions"`
	Description     string     `json:"description,omitempty"`
	Aggregation     string     `json:"aggregation,omitempty"`
	AggregationType string     `json:"result_type,omitempty"`
	OrderBy         string     `json:"order_by,omitempty"`
	GroupBy         string     `json:"group_by,omitempty"`
	DBTypes         string     `json:"db_types,omitempty"`
	Page            bool       `json:"page,omitempty"`
}

// RuleCondition represents a condition in a query rule
type RuleCondition struct {
	Expr string `json:"expr"`
}

// PrimaryKeyInfo represents primary key information
type PrimaryKeyInfo struct {
	Column string `json:"column"`
}

// ConstraintInfo represents constraint information
type ConstraintInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}

// IndexInfo represents index information
type IndexInfo struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
}
