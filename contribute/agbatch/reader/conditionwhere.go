package reader

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	cw "github.com/aif-go/ag-core/contribute/agdb/conditonwhere"
	"gorm.io/gorm"
)

var (
	paramRegex = regexp.MustCompile(`@(\w+)`)
	andOrRegex = regexp.MustCompile(`(?i)\s+(AND|OR)\s+`)
)

// FilterConditionWhere filters a template with @param placeholders, keeping only
// conditions whose @params are in the FieldMask. Returns the filtered clause.
func FilterConditionWhere(template string, fm *cw.FieldMask) string {
	if template == "" || fm == nil {
		return ""
	}
	parts := andOrRegex.Split(template, -1)
	operators := andOrRegex.FindAllString(template, -1)
	for i, op := range operators {
		operators[i] = strings.TrimSpace(op)
	}
	var kept []string
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		params := paramRegex.FindAllStringSubmatch(part, -1)
		if len(params) == 0 {
			kept = append(kept, part)
			continue
		}
		allSet := true
		for _, p := range params {
			if !fm.IsSet(p[1]) {
				allSet = false
				break
			}
		}
		if allSet {
			kept = append(kept, part)
		}
	}
	if len(kept) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString(kept[0])
	for i := 1; i < len(kept); i++ {
		op := "AND"
		if i-1 < len(operators) {
			op = operators[i-1]
		}
		b.WriteString(" ")
		b.WriteString(op)
		b.WriteString(" ")
		b.WriteString(kept[i])
	}
	return b.String()
}

// BuildConditionWhere filters a template through a FieldMask and converts @param
// placeholders to ? with ordered args, ready for GORM's Where().
func BuildConditionWhere(template string, fm *cw.FieldMask, namedArgs map[string]any) (string, []any, error) {
	if template == "" {
		return "", nil, nil
	}
	filtered := FilterConditionWhere(template, fm)
	if filtered == "" {
		return "", nil, fmt.Errorf("agbatch/reader: all conditions filtered out")
	}
	matches := paramRegex.FindAllStringSubmatch(filtered, -1)
	if len(matches) == 0 {
		return filtered, nil, nil
	}
	seen := make(map[string]bool)
	var orderedKeys []string
	for _, m := range matches {
		key := m[1]
		if !seen[key] {
			seen[key] = true
			orderedKeys = append(orderedKeys, key)
		}
	}
	args := make([]any, len(orderedKeys))
	for i, key := range orderedKeys {
		val, ok := namedArgs[key]
		if !ok {
			return "", nil, fmt.Errorf("agbatch/reader: missing named arg %q", key)
		}
		args[i] = val
	}
	return paramRegex.ReplaceAllString(filtered, "?"), args, nil
}

// GormConditionWhereQuery returns a GORM query function that uses FieldMask filtering.
func GormConditionWhereQuery(baseQuery func(*gorm.DB) *gorm.DB, template string, fm *cw.FieldMask, namedArgs map[string]any) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		where, args, err := BuildConditionWhere(template, fm, namedArgs)
		if err != nil || where == "" {
			return baseQuery(db)
		}
		return baseQuery(db).Where(where, args...)
	}
}

// GormConditionWhereQueryFunc is like GormConditionWhereQuery but evaluates the
// FieldMask at query time, allowing it to change between reads.
func GormConditionWhereQueryFunc(
	baseQuery func(*gorm.DB) *gorm.DB,
	template string,
	fmFunc func() *cw.FieldMask,
	namedArgs map[string]any,
) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		fm := fmFunc()
		where, args, err := BuildConditionWhere(template, fm, namedArgs)
		if err != nil || where == "" {
			return baseQuery(db)
		}
		return baseQuery(db).Where(where, args...)
	}
}

// MergeFieldMask merges multiple FieldMasks into one.
func MergeFieldMask(masks ...*cw.FieldMask) *cw.FieldMask {
	result := cw.NewFieldMask()
	for _, m := range masks {
		if m != nil {
			result.Merge(m)
		}
	}
	return result
}

// FieldMaskFromMap creates a FieldMask with the given keys set.
func FieldMaskFromMap(keys ...string) *cw.FieldMask {
	fm := cw.NewFieldMask()
	for _, k := range keys {
		fm.Set(k)
	}
	return fm
}

// BuildOrderedConditionWhere builds a WHERE clause with args sorted by param name.
func BuildOrderedConditionWhere(template string, fm *cw.FieldMask, namedArgs map[string]any) (string, []any, error) {
	where, args, err := BuildConditionWhere(template, fm, namedArgs)
	if err != nil {
		return "", nil, err
	}
	matches := paramRegex.FindAllStringSubmatch(where, -1)
	if len(matches) == 0 {
		return where, args, nil
	}
	type kv struct {
		key string
		idx int
	}
	var kvs []kv
	for i, m := range matches {
		kvs = append(kvs, kv{key: m[1], idx: i})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].key < kvs[j].key })
	sortedArgs := make([]any, len(kvs))
	for i, kv := range kvs {
		val, ok := namedArgs[kv.key]
		if !ok {
			return "", nil, fmt.Errorf("agbatch/reader: missing named arg %q", kv.key)
		}
		sortedArgs[i] = val
	}
	return paramRegex.ReplaceAllString(where, "?"), sortedArgs, nil
}

// Ensure imports used.
var _ = cw.FieldMask{}
var _ = gorm.Expr
var _ = fmt.Sprintf
var _ = sort.Ints
