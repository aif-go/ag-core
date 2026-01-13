package ag_conf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

const (
	BinderPlaceholderPrefix string         = "${"
	BinderPlaceholderSuffix string         = "}"
	BinderValueSeparator    string         = ":"
	KEY_BindWatched         BindWatchedKey = "BindWatched"
)

type BindWatchedKey string

var (
	ErrNotExist        = errors.New("not exist")
	ErrInvalidSyntax   = errors.New("invalid syntax")
	ErrUnBindableType  = errors.New("unbindable type")
	ErrUnsupportedType = errors.New("unsupported type")
)

var Binder IBinder

type IBinder interface {
	GetEnv() IConfigurableEnvironment
	Bind(i any, name ...string) error
	// BindValue(ctx context.Context, v reflect.Value, param BindParam) error
}

// ConfigurationPropertiesBinder 配置属性绑定器
type ConfigurationPropertiesBinder struct {
	env IConfigurableEnvironment
	// propertySources IPropertySources
}

// NewConfigurationPropertiesBinder 创建一个配置属性绑定器
func NewConfigurationPropertiesBinder(env IConfigurableEnvironment) *ConfigurationPropertiesBinder {
	cpb := &ConfigurationPropertiesBinder{}

	cpb.env = env
	// cpb.propertySources = env.GetPropertySources()

	Binder = cpb
	return cpb
}

// GetEnv 获取配置环境
func (cpb *ConfigurationPropertiesBinder) GetEnv() IConfigurableEnvironment {
	return cpb.env
}

// Bind 从指定env中绑定配置到指定的结构体
func (cpb *ConfigurationPropertiesBinder) Bind(i any, name ...string) error {
	/* - 获取反射Value，并判断是否为指针类型，并解引用 */
	var v reflect.Value
	{
		switch e := i.(type) {
		case reflect.Value:
			v = e
			if !v.IsValid() {
				return errors.New("bind value is an invalid reflect.Value")
			}
		default:
			v = reflect.ValueOf(i)
			if v.Kind() != reflect.Ptr { // 传入的绑定对象必须是指针，否则Canset为false，无法通过反射赋值
				// return unbound, errors.New("bind value should be a ptr")
				return errors.New("bind value should be a ptr")
			}
			v = v.Elem() // 获取指针指向的元素
			if !v.IsValid() {
				return errors.New("bind value points to invalid value")
			}
		}
	}

	/* - 获取反射Type，通过Type获取属性名称（配置前缀）*/
	t := v.Type() // 获取reflect.Type

	typeName := t.Name()
	if typeName == "" {
		typeName = t.String() // 基础类型的名称
	}

	rootkey := "ROOT"
	if len(name) > 0 {
		if name[0] != "" {
			rootkey = name[0]
		}
	}
	// TODO struct 中是否能通过某种方式配置prefixname

	var rootparam BindParam
	err := rootparam.BindTag(fmt.Sprintf("${%s}", rootkey), "")
	if err != nil {
		// return unbound, err
		return err
	}
	rootparam.Path = typeName

	bindcontext := context.Background()

	return cpb.BindValue(bindcontext, v, rootparam)
}

// BindValue 绑定值
func (cpb *ConfigurationPropertiesBinder) BindValue(bctx context.Context, v reflect.Value, param BindParam) (rterr error) {
	// 默认所有Binder对象都自动刷新 TODO 刷新续精确控制，添加autorefresh标签功能
	if bctx.Value(KEY_BindWatched) == nil && v.Kind() != reflect.Pointer {
		bctx = context.WithValue(bctx, KEY_BindWatched, true)
		WatcherM.RegChangeListener(param.Key, func(ck, cv string) {
			rbctx := context.WithValue(context.Background(), KEY_BindWatched, true)
			cpb.BindValue(rbctx, v, param)
		})
	}

	slog.Debug("bind value", "key", param.Key)
	defer func() {
		if rterr != nil {
			// TODO 绑定异常是否需要额外处理
		}
	}()

	if !v.CanSet() {
		err := errors.New("can not set value")
		// return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
		return fmt.Errorf("bind path=%s error: %w", param.Path, err)
	}
	// 检查Value的类型范围，只允许指定范围的类型
	if !IsBindableType(v.Type()) { // 此处的判断要保障下面代码的正确性
		slog.Error("bind value error", "key", param.Key, "type", v.Type().String(), "err", ErrUnBindableType)
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), ErrUnBindableType)
	}

	// 需要进一步解析的类型
	switch v.Kind() {
	case reflect.Pointer: // 此处的value需要解引用
		// TODO 此处v.Elem可能是空指针，会被后续!v.CanSet()判断失败。能否通过反射对nil进行提前创建实体?
		return cpb.BindValue(bctx, v.Elem(), param)
		// err := errors.New("reflect.Value shoud be ptr.Elem()")
		// return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	case reflect.Array:
		err := errors.New("use slice instead of array")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	case reflect.Slice:
		err := cpb.bindSlice(bctx, v, param)
		return err
	case reflect.Map:
		err := cpb.bindMap(bctx, v, param)
		return err
	case reflect.Struct:
		err := cpb.bindStruct(bctx, v, param)
		return err
	default:
		// do continue
	}

	return cpb.doBindValue(bctx, cpb.env, v, param)
}

func (cpb *ConfigurationPropertiesBinder) bindStruct(bctx context.Context, v reflect.Value, param BindParam) error {
	t := v.Type()

	// Struct 类型的默认值不允许有非空的默认值
	if param.PTag.HasDef && param.PTag.Def != "" {
		err := errors.New("struct can't have a non-empty default value")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	if !cpb.containsDescendantOfName(param.Key) {
		if param.PTag.HasDef {
			return nil
		}
		if param.Required {
			return fmt.Errorf("property %q %w", param.Key, ErrNotExist)
		} else {
			return nil // 中断绑定
			// TOTO 此中断是因为子字段可能还有required限制，若继续绑定其他字段，可能会导致required字段绑定失败，但此处的属性又是可选的
			// TODO 是否应该继续绑定其他字段，因为可能有默认值，若继续的话，字段可能还有required限制，怎么办？
			// TODO 后续若有自动刷新的tag支持，此处的中断存在不合理性
		}
	}

	// 遍历结构体的所有字段 // TODO 私有字段无法绑定
	for i := range t.NumField() {
		// TODO 若字段为指针，且为空指针，是否提前创建字段实例

		ft := t.Field(i) // 获取字段类型信息
		fv := v.Field(i) // 获取字段值

		// 如果字段不可fv.any()导出，跳过
		if !fv.CanInterface() {
			continue
		}

		// 创建子参数，更新路径
		subParam := BindParam{
			Key:  param.Key,
			Path: param.Path + "." + ft.Name,
		}

		// 处理匿名字段 TODO 测试场景
		if ft.Anonymous {
			// 嵌入指针类型可能导致无限递归
			if ft.Type.Kind() != reflect.Struct {
				slog.Warn(fmt.Sprintf("bind path=%s type=%s anonymous field:[%s] must be a struct", param.Path, v.Type().String(), ft.Name))
				continue
			} // 递归调用 bindStruct 方法绑定匿名结构体
			if err := cpb.bindStruct(bctx, fv, subParam); err != nil {
				return err // no wrap
			}
			continue
		}

		subParam.STag = ft.Tag

		if tag, ok := ft.Tag.Lookup("value"); ok { // 获取value标签
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return fmt.Errorf("bind path=%s type=%s error << %w", param.Path, v.Type().String(), err)
			}
		}

		if rtag, ok := ft.Tag.Lookup("required"); ok {
			// rtag 转换为bool类型
			required, err := strconv.ParseBool(rtag)
			if err != nil {
				return fmt.Errorf("bind path=%s type=%s error << %w", param.Path, v.Type().String(), err)
			}
			subParam.Required = required
		}

		// 若没有配置value标签 或 value标签设置的key为空，则使用字段名称作为key
		if subParam.Key == param.Key {
			// ft.Name 转小写
			// fname := strings.ToLower(ft.Name)
			fname := ft.Name
			// 若param.Key为空，则使用字段名称作为key
			if param.Key == "" {
				subParam.Key = fname
			} else {
				subParam.Key = fmt.Sprintf("%s.%s", param.Key, fname)
			}
		}
		// {
		// 	// 若没有配置value标签，则使用字段名称作为key
		// 	// ft.Name 转小写
		// 	// fname := strings.ToLower(ft.Name)
		// 	fname := ft.Name
		// 	subParam.Key = fmt.Sprintf("%s.%s", param.Key, fname)
		// }

		if err := cpb.BindValue(bctx, fv, subParam); err != nil {
			return err // no wrap
		}

		// TODO 若没有配置value标签，则使用字段名称作为key

	}
	return nil
}

func (cpb *ConfigurationPropertiesBinder) bindSlice(bctx context.Context, v reflect.Value, param BindParam) error {
	t := v.Type()

	et := t.Elem() // 获取切片元素类型，若t不是Array, Chan, Map, Pointer, Slice类型，会panic
	// et 可能是个指针类型
	if et.Kind() == reflect.Pointer { // 切片的子元素不能是指针
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), ErrUnBindableType)
		// et = et.Elem()
	}
	// fmt.Printf("%v\n", et.Kind())

	// 创建指定类型的新切片
	slice := reflect.MakeSlice(t, 0, 0)
	defer func() {
		v.Set(slice)
	}() // 当函数返回时，将切片设置为值 v

	for i := 0; ; i++ {
		ev := reflect.New(et).Elem()
		subParam := BindParam{
			Key:  fmt.Sprintf("%s[%d]", param.Key, i),
			Path: fmt.Sprintf("%s[%d]", param.Path, i),
		}
		if !cpb.containsDescendantOfName(subParam.Key) {
			if i == 0 && param.Required {
				// 必填时，切片类型的第一个元素不可为空
				return fmt.Errorf("bind path=%s key=%s type=%s error: %w", subParam.Path, subParam.Key, v.Type().String(), ErrNotExist)
			}
			break // 直到没有项为止
		}
		err := cpb.BindValue(bctx, ev, subParam)
		if err != nil {
			return fmt.Errorf("bind path=%s type=%s error << %w", param.Path, v.Type().String(), err)
		}
		slice = reflect.Append(slice, ev)
	}
	return nil
}

func (cpb *ConfigurationPropertiesBinder) bindMap(bctx context.Context, v reflect.Value, param BindParam) error {
	// map 类型的默认值不允许有非空的默认值
	if param.PTag.HasDef && param.PTag.Def != "" {
		err := errors.New("map can't have a non-empty default value")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	t := v.Type()
	et := t.Elem()
	ret := reflect.MakeMap(t)
	defer func() { v.Set(ret) }()

	// if !cpb.env.ContainsProperty(param.Key) {
	if !cpb.containsDescendantOfName(param.Key) {
		if param.PTag.HasDef {
			return nil
		}
		if param.Required {
			return fmt.Errorf("bind path=%s key=%s type=%s error: %w", param.Path, param.Key, v.Type().String(), ErrNotExist)
		} else {
			return nil
		}
	}

	// 1. 获取所有子键
	subKeys := cpb.getDescendantSubKeysOfName(param.Key)

	// 2. 遍历子键，构建子value
	for _, subKey := range subKeys {
		e := reflect.New(et).Elem()
		key := param.Key
		if key != "" {
			key = param.Key + "." + subKey
		}
		subParam := BindParam{
			Key:  key,
			Path: param.Path,
		}
		if err := cpb.BindValue(bctx, e, subParam); err != nil {
			return err
		}

		ret.SetMapIndex(reflect.ValueOf(subKey), e)
	}
	return nil

}

// containsDescendantOfName 检查是否包含后代键
// 子键检查包含切片类型key
// TODO 数组类型 xxx[0] 能匹配到 xxx[0]、xxx[0].key、xxx[0][1]、xxx[0][0]，应该根据具体类型精准判断是那一种情况才符合
func (cpb *ConfigurationPropertiesBinder) containsDescendantOfName(name string) bool {
	found := false
	prefix := name + "."
	prefix2 := name + "[" // 切片类型

	isArray := isArrayKey(name)

	cpb.env.GetPropertySources().RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
		// 检查属性源内容是否包含后代
		source := ps.GetSource()
		for k := range source {
			// if k == name || strings.HasPrefix(k, prefix) {
			if strings.EqualFold(k, name) || // 完全匹配
				hasPrefixIgnoreCase(k, prefix) || // .前缀匹配
				(isArray && hasPrefixIgnoreCase(k, prefix2) && isArrayKey(k)) { // 匹配二维数组
				found = true
				return true, nil
			}
		}
		return false, nil
	})

	return found
}

func isArrayKey(key string) bool {
	// 正则表达式匹配
	pattern := `^.+\[\d+\]$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(key)
}

// getDescendantKeysOfName 获取所有后代键，包含自身
// FIXME: 此处目前只有bindMap使用，暂不考虑slice的子键解析，slice在bindSlice中自行处理
func (cpb *ConfigurationPropertiesBinder) getDescendantKeysOfName(name string) []string {
	dkeys := []string{}
	prefix := name + "."
	cpb.env.GetPropertySources().RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
		// 检查属性源内容是否包含后代
		source := ps.GetSource()
		for k := range source {
			if k == name || hasPrefixIgnoreCase(k, prefix) {
				dkeys = append(dkeys, k)
			}
		}

		return false, nil // 遍历所有属性源
	})

	return dkeys
}

// getDescendantSubKeysOfName 获取所有后代子键
func (cpb *ConfigurationPropertiesBinder) getDescendantSubKeysOfName(name string) []string {
	// 获取包含name的所有键
	keys := cpb.getDescendantKeysOfName(name)
	// 解析所有键的name子键
	return getDescendantSubKeysOfName(name, keys)
}

// getDescendantSubKeysOfName 获取所有后代子键
// FIXME: 此处目前只有bindMap使用，暂不考虑slice的子键解析，slice在bindSlice中自行处理
// FIXME: bindMap情况下不存在子键起始为[的情况==
func getDescendantSubKeysOfName(name string, keys []string) []string {
	subKeys := make([]string, 0, len(keys))
	prefix := name
	if name != "" {
		prefix += "."
	}

	seen := make(map[string]bool)
	for _, k := range keys {
		if k == name { // 自身没有子键，跳过
			continue
		}

		var subKey string
		if hasPrefixIgnoreCase(k, prefix) {
			subKey = trimPrefixIgnoreCase(k, prefix)
		} else {
			continue
		}

		// 截取有效子键
		// 处理子健 1. 点分隔的情况 2.数组索引情况
		// .的索引位置
		dotIndex := strings.Index(subKey, ".") // 子键是key的情况
		// 第一个中括号的位置
		bracketIndex := strings.Index(subKey, "[") // 子键是数组的情况

		// 找到最小的有效索引
		minIndex := len(subKey)
		if dotIndex > 0 && dotIndex < minIndex {
			minIndex = dotIndex
		}
		if bracketIndex > 0 && bracketIndex < minIndex {
			minIndex = bracketIndex
		}

		// 如果找到有效索引则截取，否则保持原字符串
		if minIndex < len(subKey) {
			subKey = subKey[:minIndex]
		}

		if subKey != "" && !seen[subKey] { // 扁平的map，可能包含多个子键情况，如：[a.b.c, a.b.d] 对于a来说只有个子键：b
			seen[subKey] = true
			subKeys = append(subKeys, subKey)
		}
	}

	return subKeys
}

// /*====BindResult====*/
// // BindResult 绑定结果
// type BindResult struct {
// 	value any
// 	bound bool
// }
// // 定义一个未绑定的单例实例
// var unbound = BindResult{bound: false}
// func BindResultOf(value any) *BindResult {
// 	if reflect.ValueOf(value).IsNil() {
// 		// 由于 Go 的类型安全，我们需要进行类型转换
// 		return &BindResult{bound: false}
// 	}
// 	return &BindResult{value: value, bound: true}
// }

/*====BindParam====*/

// BindParam 绑定参数
type BindParam struct {
	Key      string // 变量对应的参数key
	Path     string // 目标变量的实际
	PTag     ParsedTag
	Required bool              // 是否必填
	STag     reflect.StructTag // 目标属性的Tag
}

func (param *BindParam) BindTag(tag string, stag reflect.StructTag) error {
	param.STag = stag
	parsedTag, err := ParseTag(tag)
	if err != nil {
		return err
	}
	if parsedTag.Key == "" { // ${:=} 默认值语法
		if parsedTag.HasDef {
			param.PTag = parsedTag
			return nil
		}
		return fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
	}
	if parsedTag.Key == "ROOT" {
		parsedTag.Key = ""
	}
	if param.Key == "" {
		param.Key = parsedTag.Key
	} else if parsedTag.Key != "" {
		param.Key = param.Key + "." + parsedTag.Key
	}
	param.PTag = parsedTag
	return nil
}

/*====ParsedTag====*/
// ParsedTag 解析后的Tag信息
type ParsedTag struct {
	Key    string // 配置名
	Def    string // 默认值
	HasDef bool   // 是否有默认值
	// Splitter string // 分割实现器名称
}

// ParseTag 解析标签字符串并返回解析后的结果
// 标签格式示例：${key:=default}>>splitter
// 其中，key 是变量名，default 是默认值（可选），splitter 是分隔符（可选）
// 返回值：
// - ret: 解析后的标签信息，包括 key、default、hasDef 和 splitter
// - err: 如果解析过程中出现错误，返回相应的错误信息
func ParseTag(tag string) (ret ParsedTag, err error) {
	// 	if tag == "" {
	// 		return ParsedTag{}, fmt.Errorf("empty tag")
	// 	}
	// 不可>>开头
	// i := strings.LastIndex(tag, ">>")
	// if i == 0 {
	// 	err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
	// 	return
	// }
	j := strings.LastIndex(tag, BinderPlaceholderSuffix)
	if j <= 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	k := strings.Index(tag, BinderPlaceholderPrefix)
	if k < 0 {
		err = fmt.Errorf("parse tag '%s' error: %w", tag, ErrInvalidSyntax)
		return
	}
	// if i > j {
	// 	ret.Splitter = strings.TrimSpace(tag[i+2:])
	// }
	ss := strings.SplitN(tag[k+2:j], BinderValueSeparator, 2)
	ret.Key = ss[0]
	if len(ss) > 1 {
		ret.HasDef = true
		ret.Def = ss[1]
	}
	return
}

// func ParseTag(tag string) (ParsedTag, error) {
// 	if tag == "" {
// 		return ParsedTag{}, fmt.Errorf("empty tag")
// 	}

// 	splitter, err := parseSplitter(tag)
// 	if err != nil {
// 		return ParsedTag{}, err
// 	}

// 	key, def, hasDef, err := parseKeyAndDefault(tag)
// 	if err != nil {
// 		return ParsedTag{}, err
// 	}

// 	return ParsedTag{
// 		Key:      key,
// 		Def:      def,
// 		HasDef:   hasDef,
// 		Splitter: splitter,
// 	}, nil
// }

// func parseSplitter(tag string) (string, error) {
// 	i := strings.LastIndex(tag, ">>")
// 	if i == 0 {
// 		return "", fmt.Errorf("invalid splitter position in tag '%s'", tag)
// 	}
// 	j := strings.LastIndex(tag, "}")
// 	if j <= 0 {
// 		return "", fmt.Errorf("missing closing brace in tag '%s'", tag)
// 	}
// 	if i > j {
// 		return strings.TrimSpace(tag[i+2:]), nil
// 	}
// 	return "", nil
// }

// func parseKeyAndDefault(tag string) (key, def string, hasDef bool, err error) {
// 	k := strings.Index(tag, "${")
// 	if k < 0 {
// 		return "", "", false, fmt.Errorf("missing variable start in tag '%s'", tag)
// 	}
// 	j := strings.LastIndex(tag, "}")
// 	if j <= 0 {
// 		return "", "", false, fmt.Errorf("missing closing brace in tag '%s'", tag)
// 	}

// 	content := tag[k+2 : j]
// 	if key, def, hasDef = strings.Cut(content, ":="); hasDef {
// 		return strings.TrimSpace(key), strings.TrimSpace(def), true, nil
// 	}
// 	return strings.TrimSpace(content), "", false, nil
// }

// IsBindableType 判断类型是否可绑定
func IsBindableType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Map, reflect.Slice:
		t = t.Elem() // 对于集合类型，检查其元素类型
	case reflect.Pointer:
		t = t.Elem() // 对于指针类型，检查其指向的类型
	case reflect.Array:
		return false // 数组类型不支持绑定，需使用Slice
	default:
		// do nothing
	}

	// 集合需要检查集合的元素类型
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return true
	case reflect.Float32, reflect.Float64:
		return true
	case reflect.Bool:
		return true
	case reflect.String:
		return true
	case reflect.Struct:
		return true
	case reflect.Map, reflect.Slice: // 集合元素可以是集合
		return true
	case reflect.Pointer:
		// return IsBindableType(t.Elem()) // FIXME 是否允许集合元素为指针类型，当前不支持
		return false
	default:
		return false
	}
}

// doBindValue 绑定配置值到目标结构体字段
func (cpb *ConfigurationPropertiesBinder) doBindValue(ctx context.Context, env IConfigurableEnvironment, v reflect.Value, param BindParam) error {
	// 需要获取参数的类型
	value := env.GetProperty(param.Key) // TODO value中的占位符按设计需要在env中完成解析，是否需要在此处处理？
	// cpb.env.ContainsProperty(param.Key)

	if value == "" {
		// 没有找到配置，则使用默认值
		if param.PTag.HasDef {
			// TODO 告警日志，使用默认值
			slog.Warn(fmt.Sprintf("bind path=%s type=%s use default value=%s\n", param.Path, v.Type().String(), param.PTag.Def))
			value = param.PTag.Def
		} else if param.Required {
			// 根据必填标志判断是否报错
			return fmt.Errorf("bind path=%s key=%s type=%s error: %w", param.Path, param.Key, v.Type().String(), ErrNotExist)
		} else {
			return nil // 非必填字段，且没有配置默认值，直接返回
		}
	}

	// TODO 默认值可能也有占位符 默认值暂不支持占位符
	return parseValue(value, v, param)
}

// parseValue 将string类型的value按照reflect.Value的类型进行转换，并赋值给v
func parseValue(value string, v reflect.Value, param BindParam) error {
	// 检查目标值是否可设置，避免panic
	if !v.CanSet() {
		// return fmt.Errorf("bind path=%s type=%s error: value is not settable", param.Path, v.Type().String())
		// FIXME v.Type().String()空指针会panic
		return fmt.Errorf("bind path=%s error: value is not settable", param.Path)
	}

	// 封装错误处理，减少重复代码
	wrapError := func(err error) error {
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	switch v.Kind() {
	case reflect.Int:
		i, err := strconv.ParseInt(value, 0, 64) // int位数与平台相关，用64兼容大部分场景
		if err != nil {
			return wrapError(err)
		}
		v.SetInt(i)
	case reflect.Int8:
		i, err := strconv.ParseInt(value, 0, 8)
		if err != nil {
			return wrapError(err)
		}
		v.SetInt(i)
	case reflect.Int16:
		i, err := strconv.ParseInt(value, 0, 16)
		if err != nil {
			return wrapError(err)
		}
		v.SetInt(i)
	case reflect.Int32:
		i, err := strconv.ParseInt(value, 0, 32)
		if err != nil {
			return wrapError(err)
		}
		v.SetInt(i)
	case reflect.Int64:
		i, err := strconv.ParseInt(value, 0, 64)
		if err != nil {
			return wrapError(err)
		}
		v.SetInt(i)

	case reflect.Uint:
		i, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return wrapError(err)
		}
		v.SetUint(i)
	case reflect.Uint8:
		i, err := strconv.ParseUint(value, 0, 8)
		if err != nil {
			return wrapError(err)
		}
		v.SetUint(i)
	case reflect.Uint16:
		i, err := strconv.ParseUint(value, 0, 16)
		if err != nil {
			return wrapError(err)
		}
		v.SetUint(i)
	case reflect.Uint32:
		i, err := strconv.ParseUint(value, 0, 32)
		if err != nil {
			return wrapError(err)
		}
		v.SetUint(i)
	case reflect.Uint64:
		i, err := strconv.ParseUint(value, 0, 64)
		if err != nil {
			return wrapError(err)
		}
		v.SetUint(i)

	case reflect.Float32:
		f, err := strconv.ParseFloat(value, 32)
		if err != nil {
			return wrapError(err)
		}
		v.SetFloat(f)
	case reflect.Float64:
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return wrapError(err)
		}
		v.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return wrapError(err)
		}
		v.SetBool(b)

	case reflect.String:
		v.SetString(value)

	default:
		return wrapError(ErrUnsupportedType)
	}

	return nil
}
