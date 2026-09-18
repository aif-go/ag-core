package model

import (
	"fmt"
	"strings"

	"github.com/aif-go/ag-core/tool/cmd/gen-go-db/table"
)

// hasWhereDataToYAMLCache 检查是否有 WhereDataToYAMLCache 需要生成
func hasWhereDataToYAMLCache(tableData *table.TableData) bool {
	if len(tableData.SelfQueries) == 0 {
		return false
	}
	for _, query := range tableData.SelfQueries {
		if query.WhereDataYaml != "" {
			return true
		}
	}
	return false
}

// generateImports 生成导入语句
func generateImports(tableData *table.TableData) string {
	var imports []string

	// 基础导入
	imports = append(imports, "\"fmt\"")

	// 根据需要添加导入
	if tableData.HasPage {
		imports = append(imports, `db "github.com/aif-go/ag-core/contribute/agdb/gormdb"`)
	}
	if tableData.HasSelfQuery {
		imports = append(imports, `"github.com/aif-go/ag-core/contribute/agdb/conditonwhere"`)
	}

	// 如果有 WhereDataToYAMLCache，需要导入 yaml 包
	if hasWhereDataToYAMLCache(tableData) {
		imports = append(imports, `"gopkg.in/yaml.v2"`)
	}

	// 添加自定义导入包
	if tableData.ModelTemplateData != nil && len(tableData.ModelTemplateData.ImportPackages) > 0 {
		for _, pkg := range tableData.ModelTemplateData.ImportPackages {
			if pkg != "fmt" { // 避免重复
				imports = append(imports, fmt.Sprintf("\"%s\"", pkg))
			}
		}
	}

	if len(imports) == 0 {
		return ""
	}

	return "import (\n\t" + strings.Join(imports, "\n\t") + "\n)"
}

// generateStructFields 生成主结构体字段
func generateStructFields(tableData *table.TableData) string {
	var fields []string

	for _, col := range tableData.Columns {
		fields = append(fields, fmt.Sprintf("\t%s %s `gorm:\"%s\" json:\"%s\"`",
			col.JsonTag, col.GoType, col.GormTag, col.JsonTag))
	}

	return strings.Join(fields, "\n")
}

// generatePrimaryKeyTypeBlock 生成主键类型定义块
func generatePrimaryKeyTypeBlock(tableData *table.TableData) string {
	structName := tableData.StructName

	if len(tableData.PrimaryKeys) == 1 {
		// 单主键，使用类型别名
		for _, col := range tableData.Columns {
			if col.Name == tableData.PrimaryKeys[0] {
				return fmt.Sprintf(`// %sPrimaryKey 单主键类型别名
type %sPrimaryKey %s`, structName, structName, col.GoType)
			}
		}
	}

	// 多主键，使用结构体
	var fields []string
	for _, pk := range tableData.PrimaryKeys {
		for _, col := range tableData.Columns {
			if col.Name == pk {
				fields = append(fields, fmt.Sprintf("\t%s %s", col.JsonTag, col.GoType))
				break
			}
		}
	}

	return fmt.Sprintf(`// %sPrimaryKey 多主键结构体
type %sPrimaryKey struct {
%s
}

// %sPrimarkey 多主键结构体（历史拼写别名，永久保留，无移除计划）
type %sPrimarkey = %sPrimaryKey`, structName, structName, strings.Join(fields, "\n"), structName, structName, structName)
}

// lowerStructName 首字母小写的结构体名。
func lowerStructName(structName string) string {
	return strings.ToLower(string(structName[0])) + structName[1:]
}

// generateCloneFields 生成Clone方法的字段列表
func generateCloneFields(tableData *table.TableData) string {
	lsn := lowerStructName(tableData.StructName)
	var fields []string
	for _, col := range tableData.Columns {
		fields = append(fields, fmt.Sprintf("\t\t%s: %s.%s,", col.JsonTag, lsn, col.JsonTag))
	}
	return strings.Join(fields, "\n")
}

// isIndexColumn 检查列是否为主键或索引列
func isIndexColumn(tableData *table.TableData, colName string) bool {
	// 检查是否为主键
	for _, pk := range tableData.PrimaryKeys {
		if pk == colName {
			return true
		}
	}
	// 检查是否为索引列
	for _, index := range tableData.Indexes {
		for _, idxCol := range index.Columns {
			if idxCol == colName {
				return true
			}
		}
	}
	return false
}

// isSpecialColumn 检查列是否为特殊列（jpaVersion, create_time, last_update_time）
func isSpecialColumn(colData table.ColumnData) bool {
	if colData.IsAutoCreate || colData.IsAutoUpdate || colData.IsJavaVersion {
		return true
	}
	return false
}

// getZeroCheck 获取零值检查表达式（参考dao模板的简洁实现）
func getZeroCheck(lowerName string, col table.ColumnData) string {
	// 指针类型列使用 nil 判断
	if strings.HasPrefix(col.GoType, "*") {
		return fmt.Sprintf("%s.%s == nil", lowerName, col.JsonTag)
	}
	switch col.GoType {
	case "string":
		return fmt.Sprintf("%s.%s == \"\"", lowerName, col.JsonTag)
	case "time.Time", "decimal.Decimal":
		return fmt.Sprintf("%s.%s.IsZero()", lowerName, col.JsonTag)
	case "bool":
		return fmt.Sprintf("!%s.%s", lowerName, col.JsonTag)
	default:
		// 数值类型（int, int64, float64等）
		return fmt.Sprintf("%s.%s == 0", lowerName, col.JsonTag)
	}
}

// containsPk 检查列名是否在主键列表中
func containsPk(primaryKeys []string, colName string) bool {
	for _, pk := range primaryKeys {
		if pk == colName {
			return true
		}
	}
	return false
}

// FieldCheck 字段零值检查数据（§7.2）。
type FieldCheck struct {
	Tag      string
	ColName  string
	ZeroCond string
	Kind     string // primary / index / special / general
	Desc     string // 特殊列用途描述
}

// buildFieldChecks 构建字段检查数据。
func buildFieldChecks(tableData *table.TableData) []FieldCheck {
	var checks []FieldCheck
	lsn := lowerStructName(tableData.StructName)
	for _, col := range tableData.Columns {
		isPrimary := containsPk(tableData.PrimaryKeys, col.Name)
		isIndex := isIndexColumn(tableData, col.Name)
		isSpecial := isSpecialColumn(col)
		check := FieldCheck{
			Tag:      col.JsonTag,
			ColName:  col.Name,
			ZeroCond: getZeroCheck(lsn, col),
		}
		switch {
		case isPrimary:
			check.Kind = "primary"
		case isIndex:
			check.Kind = "index"
		case isSpecial:
			check.Kind = "special"
			check.Desc = "特殊用途"
			if strings.Contains(strings.ToLower(col.Name), "version") {
				check.Desc = "乐观锁"
			} else if strings.Contains(strings.ToLower(col.Name), "create") {
				check.Desc = "自动创建时间"
			} else if strings.Contains(strings.ToLower(col.Name), "update") {
				check.Desc = "自动更新时间"
			}
		default:
			check.Kind = "general"
		}
		checks = append(checks, check)
	}
	return checks
}

// HasGeneralCol 是否存在普通列（决定 generalColZeroVal 声明）。
func (d *ModelTemplateData) HasGeneralCol() bool {
	for _, c := range d.FieldChecks {
		if c.Kind == "general" {
			return true
		}
	}
	return false
}


// LowerStructName 首字母小写结构体名（模板用）。
func (d *ModelTemplateData) LowerStructName() string {
	return lowerStructName(d.StructName)
}
// FieldChecksCode 渲染字段检查代码（原 generateListZeroValueColsMethod 内部逻辑）。
func (d *ModelTemplateData) FieldChecksCode() string {
	lsn := lowerStructName(d.StructName)
	var fieldChecks []string

	// 先遍历判断是否存在普通列，有才声明 generalColZeroVal
	if d.HasGeneralCol() {
		generalColZeroValVarDefine := fmt.Sprintf(` // generalColZeroVal 用于普通列的零值检查，避免重复代码
	%s
	`, "var generalColZeroVal bool = false")
		fieldChecks = append(fieldChecks, generalColZeroValVarDefine)
	}

	for _, fc := range d.FieldChecks {
		zeroCheck := fc.ZeroCond
		var fieldCheck string
		switch fc.Kind {
		case "primary":
			fieldCheck = fmt.Sprintf(`	// %s - 主键，索引列
	if !filterPrimary {
		isZero := %s
		// false 保留零值 true 过滤零值
		if (!filterIsZero && isZero) || (filterIsZero && !isZero) {
			cols = append(cols, "%s")
			vals = append(vals, %s.%s)
		}
	}`, fc.Tag, zeroCheck, fc.ColName, lsn, fc.Tag)
		case "index":
			fieldCheck = fmt.Sprintf(`	// %s - 索引列
	if !filterIndex {
		isZero := %s
		if (!filterIsZero && isZero) || (filterIsZero && !isZero) {
			cols = append(cols, "%s")
			vals = append(vals, %s.%s)
		}
	}`, fc.Tag, zeroCheck, fc.ColName, lsn, fc.Tag)
		case "special":
			fieldCheck = fmt.Sprintf(`	// %s - 特殊列，用于%s
	if !filterSpecial {
		isZero := %s
		if (!filterIsZero && isZero) || (filterIsZero && !isZero) {
			cols = append(cols, "%s")
			vals = append(vals, %s.%s)
		}
	}`, fc.Tag, fc.Desc, zeroCheck, fc.ColName, lsn, fc.Tag)
		default:
			fieldCheck = fmt.Sprintf(`	// %s - 普通列
	generalColZeroVal = %s
	if (!filterIsZero && generalColZeroVal) || (filterIsZero && !generalColZeroVal) {
		cols = append(cols, "%s")
		vals = append(vals, %s.%s)
	}`, fc.Tag, zeroCheck, fc.ColName, lsn, fc.Tag)
		}
		fieldChecks = append(fieldChecks, fieldCheck)
	}
	return strings.Join(fieldChecks, "\n\n")
}

// generateAllowUpdateColsBlock 生成AllowUpdateCols变量块
func generateAllowUpdateColsBlock(tableData *table.TableData) string {
	structName := tableData.StructName

	var cols []string
	for i, col := range tableData.AllowUpdateCols {
		if i > 0 {
			cols = append(cols, fmt.Sprintf(", \"%s\"", col))
		} else {
			cols = append(cols, fmt.Sprintf("\"%s\"", col))
		}
	}

	return fmt.Sprintf(`// %sAllowUpdateCols 支持更新的列名列表
var %sAllowUpdateCols = []string{%s}`, structName, structName, strings.Join(cols, ""))
}

// generateIndexLeadingColsBlock 生成IndexLeadingCols变量块
func generateIndexLeadingColsBlock(tableData *table.TableData) string {
	structName := tableData.StructName

	var cols []string
	// 优先使用主键的第一列
	if len(tableData.PrimaryKeys) > 0 {
		cols = append(cols, fmt.Sprintf("\"%s\"", tableData.PrimaryKeys[0]))
	}
	// 处理索引的引导列
	for _, index := range tableData.Indexes {
		if len(index.Columns) > 0 {
			cols = append(cols, fmt.Sprintf("\"%s\"", index.Columns[0]))
		}
	}
	if len(cols) == 0 {
		return ""
	}

	return fmt.Sprintf(`// %sIndexLeadingCols 索引前导列，用于检查是否走了索引，避免全表扫描
// 列名必须和数据库表列名一致，区分大小写
var %sIndexLeadingCols = []string{%s}`, structName, structName, strings.Join(cols, ", "))
}

// generateQueryArgStruct 生成查询参数结构体
func generateQueryArgStruct(tableData *table.TableData, query table.QueryData) string {
	structName := tableData.StructName
	queryName := query.Name

	var fields []string

	if query.HasPage {
		fields = append(fields, "\tdb.Page")
		fields = append(fields, "\tFieldMask *conditonwhere.FieldMask")
	} else {
		fields = append(fields, "\tFieldMask *conditonwhere.FieldMask")
	}

	for _, wc := range query.WhereColFields {
		// 如果前期已经指定了参数类型，则按照指定的处理
		if wc.GoType != "" {
			if wc.IsSlice {
				fields = append(fields, fmt.Sprintf("\t%s []%s", wc.FieldName, wc.GoType))
			} else {
				fields = append(fields, fmt.Sprintf("\t%s %s", wc.FieldName, wc.GoType))
			}
			continue
		}
		for _, col := range tableData.Columns {
			if col.Name == wc.ColName {
				if wc.IsSlice {
					fields = append(fields, fmt.Sprintf("\t%s []%s", wc.FieldName, col.GoType))
				} else {
					fields = append(fields, fmt.Sprintf("\t%s %s", wc.FieldName, col.GoType))
				}
				break
			}
		}
	}

	return fmt.Sprintf(`// %s%sArg %s 查询参数
type %s%sArg struct {
%s
}`, structName, queryName, queryName, structName, queryName, strings.Join(fields, "\n"))
}

// generateWithMethods 生成With方法
func generateWithMethods(tableData *table.TableData, query table.QueryData) string {
	structName := tableData.StructName
	queryName := query.Name
	lsn := lowerStructName(structName)

	var methods []string

	for _, wc := range query.WhereColFields {
		// 对于用户指定的go类型，此处只需按照指定的类型处理
		if wc.GoType != "" {
			// CHG-07：切片参数（in/not in）生成 []T 形参，与 Arg 结构体字段类型一致
			goType := wc.GoType
			if wc.IsSlice {
				goType = "[]" + goType
			}
			methods = append(methods, fmt.Sprintf(`func (%s%sArg *%s%sArg) With%s(%s %s) *%s%sArg{
	%s%sArg.%s = %s
	%s%sArg.FieldMask.Set("%s")
	return %s%sArg
} `, lsn, queryName, structName, queryName, wc.FieldName, wc.FieldName, goType, structName, queryName,
				lsn, queryName, wc.FieldName, wc.FieldName, lsn, queryName, wc.FieldName, lsn, queryName))
			continue
		}
		for _, col := range tableData.Columns {
			if col.Name == wc.ColName {
				// CHG-07：切片参数（in/not in）生成 []T 形参，与 Arg 结构体字段类型一致
				goType := col.GoType
				if wc.IsSlice {
					goType = "[]" + goType
				}
				methods = append(methods, fmt.Sprintf(`func (%s%sArg *%s%sArg) With%s(%s %s) *%s%sArg{
	%s%sArg.%s = %s
	%s%sArg.FieldMask.Set("%s")
	return %s%sArg
} `, lsn, queryName, structName, queryName, wc.FieldName, wc.FieldName, goType, structName, queryName,
					lsn, queryName, wc.FieldName, wc.FieldName, lsn, queryName, wc.FieldName, lsn, queryName))
				break
			}
		}
	}

	return strings.Join(methods, "\n\n")
}

// generateConvertToMapMethod 生成ConvertToMap方法
func generateConvertToMapMethod(tableData *table.TableData, query table.QueryData) string {
	structName := tableData.StructName
	queryName := query.Name
	lsn := lowerStructName(structName)

	var fields []string

	if query.HasPage {
		fields = append(fields, "\t\t\"Page\": "+lsn+queryName+"Arg.Page,")
	}

	for _, wc := range query.WhereColFields {
		fields = append(fields, fmt.Sprintf("\t\t\"%s\": %s%sArg.%s,", wc.FieldName, lsn, queryName, wc.FieldName))
	}

	return fmt.Sprintf(`// ConvertToMap 将参数转换为map
func (%s%sArg *%s%sArg) ConvertToMap() map[string]interface{} {
	if %s%sArg == nil {
		return nil
	}
	return map[string]interface{}{
%s
	}
}`, lsn, queryName, structName, queryName, lsn, queryName, strings.Join(fields, "\n"))
}

// generateQueryResStruct 生成查询结果结构体
func generateQueryResStruct(tableData *table.TableData, query table.QueryData) string {
	structName := tableData.StructName
	queryName := query.Name

	if query.SelectFields == "*" {
		// 使用主结构体
		return ""
	}

	var fields []string
	for _, col := range tableData.Columns {
		if contains(query.Fields, col.Name) {
			fields = append(fields, fmt.Sprintf("\t%s %s `gorm:\"column:%s\" json:\"%s\"`",
				col.JsonTag, col.GoType, col.Name, col.JsonTag))
		}
	}

	return fmt.Sprintf(`// %s%sRes %s 查询结果
type %s%sRes struct {
%s
}`, structName, queryName, queryName, structName, queryName, strings.Join(fields, "\n"))
}

// contains 检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generatePageResStruct 生成分页结果结构体
func generatePageResStruct(tableData *table.TableData, query table.QueryData) string {
	structName := tableData.StructName
	queryName := query.Name

	var resultType string
	if query.SelectFields == "*" {
		resultType = structName
	} else {
		resultType = structName + queryName + "Res"
	}

	return fmt.Sprintf(`// %s%sPageRes %s 分页查询结果
type %s%sPageRes struct {
	db.PageResult
	ResultList []*%s
}`, structName, queryName, queryName, structName, queryName, resultType)
}

// QueryBlock 单个自定义查询代码块（§7.2）。
type QueryBlock struct {
	Name       string
	Code       string
	HasPage    bool
	HasRes     bool
	ResultType string
}

// buildQueryBlocks 逐查询组装代码块（原 generateQueryCode 逻辑）。
func buildQueryBlocks(tableData *table.TableData) []QueryBlock {
	var blocks []QueryBlock
	for _, query := range tableData.SelfQueries {
		var code string
		code += generateQueryArgStruct(tableData, query)
		code += "\n\n"
		code += generateWithMethods(tableData, query)
		code += "\n\n"
		code += generateConvertToMapMethod(tableData, query)
		code += "\n\n"
		hasRes := query.SelectFields != "*"
		if hasRes {
			code += generateQueryResStruct(tableData, query)
			code += "\n\n"
		}
		if query.HasPage {
			code += generatePageResStruct(tableData, query)
			code += "\n\n"
		}
		var resultType string
		if query.SelectFields == "*" {
			resultType = tableData.StructName
		} else {
			resultType = tableData.StructName + query.Name + "Res"
		}
		blocks = append(blocks, QueryBlock{
			Name:       query.Name,
			Code:       code,
			HasPage:    query.HasPage,
			HasRes:     hasRes,
			ResultType: resultType,
		})
	}
	return blocks
}

// generateWhereDataToYAMLCache 生成WhereDataToYAMLCache map
func generateWhereDataToYAMLCache(tableData *table.TableData) string {
	if len(tableData.SelfQueries) == 0 {
		return ""
	}

	var entries []string

	for _, query := range tableData.SelfQueries {
		if query.WhereDataYaml != "" {
			key := query.Name
			entries = append(entries, fmt.Sprintf("\t\"%s\":`%s`,", key, query.WhereDataYaml))
		}
	}

	if len(entries) == 0 {
		return ""
	}

	return fmt.Sprintf(`// WhereDataToYAMLCache WhereDataYAML缓存
var `+tableData.StructName+`WhereDataToYAMLCache = map[string]string{
%s
}`, strings.Join(entries, "\n"))
}

// generateInitFunctionBody 生成 init 函数用于注册自定义 condition
func generateInitFunctionBody(tableData *table.TableData) string {
	if !hasWhereDataToYAMLCache(tableData) {
		return ""
	}

	return `
var ` + tableData.StructName + `ConditionMap = map[string]*conditonwhere.MaskWhereCondition{}

// 初始注册自定义函数的condition
func init() {
	for key, whereDataYaml := range ` + tableData.StructName + `WhereDataToYAMLCache {
		var newData map[interface{}]interface{}
		yaml.Unmarshal([]byte(whereDataYaml), &newData)
		 ` + tableData.StructName + `ConditionMap[key] = conditonwhere.ParseWhereCondition(newData)
	}
}
`
}

// ModelTemplateData model 模板渲染数据（§7.2）。
type ModelTemplateData struct {
	*table.TableData
	ImportsBlock          string
	StructFields          string
	PrimaryKeyTypeBlock   string
	CloneFields           string
	AllowUpdateColsBlock  string
	IndexLeadingColsBlock string
	FieldChecks           []FieldCheck
	QueryBlocks           []QueryBlock
	HasWhereCache         bool
	whereCacheBlock       string
	initFuncBlock         string
}

// WhereCacheBlock WhereDataToYAMLCache 代码块（含前导分隔，无则空）。
func (d *ModelTemplateData) WhereCacheBlock() string { return d.whereCacheBlock }

// InitFuncBlock condition 注册 init 代码块（含前导分隔，无则空）。
func (d *ModelTemplateData) InitFuncBlock() string { return d.initFuncBlock }

// BuildModelTemplateData 构建 model 模板渲染数据。
func BuildModelTemplateData(tableData *table.TableData) *ModelTemplateData {
	data := &ModelTemplateData{
		TableData:             tableData,
		ImportsBlock:          generateImports(tableData),
		StructFields:          generateStructFields(tableData),
		PrimaryKeyTypeBlock:   generatePrimaryKeyTypeBlock(tableData),
		CloneFields:           generateCloneFields(tableData),
		AllowUpdateColsBlock:  generateAllowUpdateColsBlock(tableData),
		IndexLeadingColsBlock: generateIndexLeadingColsBlock(tableData),
		FieldChecks:           buildFieldChecks(tableData),
		QueryBlocks:           buildQueryBlocks(tableData),
		HasWhereCache:         hasWhereDataToYAMLCache(tableData),
	}
	if cache := generateWhereDataToYAMLCache(tableData); cache != "" {
		data.whereCacheBlock = "\n\n" + cache
	}
	if init := generateInitFunctionBody(tableData); init != "" {
		data.initFuncBlock = "\n\n" + init
	}
	return data
}
