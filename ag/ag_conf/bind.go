package ag_conf

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

const (
	BinderPlaceholderPrefix string = "${"
	BinderPlaceholderSuffix string = "}"
	BinderValueSeparator    string = ":"
)

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
	slog.Debug("bind value", "key", param.Key)
	defer func() {
		if rterr != nil {
			// TODO 绑定异常是否需要额外处理
		}
	}()

	if !v.CanSet() {
		err := errors.New("can not set value")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}
	// 检查Value的类型范围，只允许指定范围的类型
	if !IsBindableType(v.Type()) { // 此处的判断要保障下面代码的正确性
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), ErrUnBindableType)
	}

	// 需要进一步解析的类型
	switch v.Kind() {
	case reflect.Pointer: // 此处的value需要解引用
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

	// 遍历结构体的所有字段 // TODO 私有字段无法绑定
	for i := range t.NumField() {
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

		if tag, ok := ft.Tag.Lookup("value"); ok { // 获取value标签
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return fmt.Errorf("bind path=%s type=%s error << %w", param.Path, v.Type().String(), err)
			}
		}

		// 若没有配置value标签 或 value标签设置的key为空，则使用字段名称作为key
		if subParam.Key == param.Key {
			// ft.Name 转小写
			// fname := strings.ToLower(ft.Name)
			fname := ft.Name
			subParam.Key = fmt.Sprintf("%s.%s", param.Key, fname)
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
	// TODO 如果为直接属性 且 param autorefresh 添加自动参数刷新
	// vsetfunc := func() {
	// 	vsctx := context.WithValue(context.Background(), "canautorefresh", false)
	// 	cpb.bindSlice(vsctx, v, param)
	// }

	t := v.Type()

	et := t.Elem() // 获取切片元素类型，若t不是Array, Chan, Map, Pointer, Slice类型，会panic
	// fmt.Printf("%v\n", et.Kind())

	// 创建指定类型的新切片
	slice := reflect.MakeSlice(t, 0, 0)
	defer func() {
		v.Set(slice)
	}() // 当函数返回时，将切片设置为值 v
	// ============ TEST ==========

	// for i := 0; i < 5; i++ {
	// 	ev := reflect.New(et).Elem()
	// 	ev.FieldByName("String").SetString(fmt.Sprintf("value-%d", i))
	// 	slice = reflect.Append(slice, ev)
	// }
	// ============ TEST END =======

	for i := 0; ; i++ {
		ev := reflect.New(et).Elem()
		subParam := BindParam{
			Key:  fmt.Sprintf("%s[%d]", param.Key, i),
			Path: fmt.Sprintf("%s[%d]", param.Path, i),
		}
		// subParam.BindTag("tag string", "")

		// TODO 判断下标不存在则中断循环
		// if !cpb.env.ContainsProperty(subParam.Key) { // TODO 暂未实现a.b[0].c:123 情况的判断
		if !cpb.containsDescendantOfName(subParam.Key) {
			break
		}
		err := cpb.BindValue(bctx, ev, subParam) // TODO 待优化
		// if errors.Is(err, ErrNotExist) {   // 按此处未找到判断，不严谨，若切片类型为struct，配置时，某个元素的属性确实没配置，此处存在误处理的情况
		// 	break
		// }
		if err != nil {
			return fmt.Errorf("bind path=%s type=%s error << %w", param.Path, v.Type().String(), err)
			// return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
		}
		slice = reflect.Append(slice, ev)
	}
	return nil
}

func (cpb *ConfigurationPropertiesBinder) bindMap(bctx context.Context, v reflect.Value, param BindParam) error {
	if param.PTag.HasDef && param.PTag.Def != "" {
		err := errors.New("map can't have a non-empty default value")
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
	}

	t := v.Type()
	et := t.Elem()
	ret := reflect.MakeMap(t)
	defer func() { v.Set(ret) }()

	if param.PTag.Key == "" {
		if param.PTag.HasDef {
			return nil
		}
		return fmt.Errorf("tag for %s requires a default value", param.Path)
	}

	// if !cpb.env.ContainsProperty(param.Key) {
	if !cpb.containsDescendantOfName(param.Key) {
		if param.PTag.HasDef {
			return nil
		}
		return fmt.Errorf("property %q %w", param.Key, ErrNotExist)
	}

	// 1. 获取所有子键
	// keys := cpb.getDescendantKeysOfName(param.Key)
	keys := cpb.getDescendantSubKeysOfName(param.Key)
	// keys := make([]string, 0)
	// keys = append(keys, "hzw1")
	// keys = append(keys, "hzw2")
	// keys, err := p.SubKeys(param.Key)
	// if err != nil {
	// 	return errutil.WrapError(err, "bind path=%s type=%s error", param.Path, v.Type().String())
	// }

	// 2. 遍历子键，构建子value
	for _, key := range keys {
		e := reflect.New(et).Elem()
		subKey := key
		if param.Key != "" {
			subKey = param.Key + "." + key
		}
		subParam := BindParam{
			Key:  subKey,
			Path: param.Path,
		}
		if err := cpb.BindValue(bctx, e, subParam); err != nil {
			return err
		}

		// if err = BindValue(p, e, et, subParam, filter); err != nil {
		// 	return err // no wrap
		// }
		ret.SetMapIndex(reflect.ValueOf(key), e)
	}
	return nil

}

func (cpb *ConfigurationPropertiesBinder) containsDescendantOfName(name string) bool {
	found := false
	prefix := name + "."

	// cpb.propertySources.RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
	cpb.env.GetPropertySources().RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
		// 检查属性源内容是否包含后代
		source := ps.GetSource()
		for k := range source {
			if k == name || strings.HasPrefix(k, prefix) {
				found = true
				return true, nil
			}
		}
		return false, nil
	})

	return found
}

func (cpb *ConfigurationPropertiesBinder) getDescendantKeysOfName(name string) []string {
	dkeys := []string{}
	prefix := name + "."
	// cpb.propertySources.RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
	cpb.env.GetPropertySources().RangePropertySourceHandler(func(ps IPropertySource) (end bool, err error) {
		// 检查属性源内容是否包含后代
		source := ps.GetSource()
		for k := range source {
			if k == name || strings.HasPrefix(k, prefix) {
				dkeys = append(dkeys, k)
			}
		}

		return false, nil // 遍历所有属性源
	})

	return dkeys
}

func (cpb *ConfigurationPropertiesBinder) getDescendantSubKeysOfName(name string) []string {
	keys := cpb.getDescendantKeysOfName(name)
	subKeys := make([]string, 0, len(keys))
	prefix := name
	if prefix != "" {
		prefix += "."
	}

	seen := make(map[string]bool)
	for _, k := range keys {
		if k == name {
			continue
		}
		subKey := strings.TrimPrefix(k, prefix)
		// 只取第一级子键
		if dot := strings.Index(subKey, "."); dot > 0 {
			subKey = subKey[:dot]
		}
		if !seen[subKey] {
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
	Key  string // 变量对应的参数key
	Path string // 目标变量的实际
	PTag ParsedTag
	STag reflect.StructTag // 目标属性的Tag
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
	case reflect.Pointer:
		// return IsBindableType(t.Elem()) // TODO 是否允许集合元素为指针类型
		return false
	default:
		return false
	}
}

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
		} else {
			return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), ErrNotExist)
		}
	}
	// TODO 默认值可能也有占位符 默认值暂不支持占位符

	// 将string 类型的value，按照reflect.Value的类型进行转换，并赋值给v
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 解析为int类型
		if i, err := strconv.ParseInt(value, 0, 0); err == nil {
			v.SetInt(i)
			return nil
		} else {
			return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 解析为uint类型
		if i, err := strconv.ParseUint(value, 0, 0); err == nil {
			v.SetUint(i)
			return nil
		} else {
			return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
		}
	case reflect.Float32, reflect.Float64:
		// 解析为float类型
		if f, err := strconv.ParseFloat(value, 0); err == nil {
			v.SetFloat(f)
			return nil
		} else {
			return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
		}

	case reflect.Bool:
		// 解析为bool类型
		if b, err := cast.ToBoolE(value); err == nil {
			// if b, err := strconv.ParseBool(value); err == nil { // TODO
			v.SetBool(b)
			return nil
		} else {
			return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)
		}
	case reflect.String:
		// 解析为string类型
		v.SetString(value)
		return nil
	default:
		// 其他类型无法解析
		// err := errors.New("unsupported type")
		err := ErrUnsupportedType
		return fmt.Errorf("bind path=%s type=%s error: %w", param.Path, v.Type().String(), err)

	}

}
