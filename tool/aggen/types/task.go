package types

import (
	"bytes"
	"text/template"
)

// OutputType 输出类型
type OutputType int

const (
	// OutputTypeOverride 覆盖
	OutputTypeOverride OutputType = iota

	// OutputTypeNotOverride 不覆盖
	OutputTypeNotOverride

	// OutputTypeAppend 追加
	OutputTypeAppend // TODO 未实现

	// OutputTypeInsert 插入，实现在指定位置插入内容
	OutputTypeInsert // TODO 未实现
)

type ScopeType string

const (
	ScopeModule       ScopeType = "module"
	ScopePackage      ScopeType = "package"
	ScopePackageGroup ScopeType = "package_group"
	ScopeService      ScopeType = "service"
)

// File .
type File struct {
	// 文件名（路径 + 文件名）
	Name string
	// 文件内容
	Content string

	// FIXME ext by houzw
	OutputType OutputType
}

// Task .
type Task struct {
	Name      string                      // 文件名
	Path      string                      // 文件路径，包含文件名
	Text      string                      // 文件内容
	SetImport func(pkg *GennerInfo) error // 设置导入包, 每个文件的import都是不一样的，应该在定义task时指定

	*template.Template // 模板

	// FIXME ext by houzw
	Scope      ScopeType  // 作用域，module、package、service
	OutputType OutputType // 输出类型，覆盖、追加、插入
}


// Build .
func (t *Task) Build() error {
	x, err := template.New(t.Name).Funcs(funcs).Parse(t.Text)
	if err != nil {
		return err
	}

	t.Template = x
	return nil
}

// Render 渲染任务.
func (t *Task) Render(data interface{}) (*File, error) {
	var buf bytes.Buffer
	if t.Text != "" {
		if t.Template == nil {
			err := t.Build()
			if err != nil {
				return nil, err
			}
		}
		err := t.ExecuteTemplate(&buf, t.Name, data)
		if err != nil {
			return nil, err
		}

	}

	// if t.RenderGen != nil {
	// 	text, err := t.RenderGen(data)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	_, err = buf.Write([]byte(text))
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return &File{
		Name:       t.Path,
		Content:    buf.String(),
		OutputType: t.OutputType,
	}, nil
}
