package render

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"sync"
)

//go:embed templates/*.tmpl
var TemplateFS embed.FS

// 模板缓存映射，用于缓存已加载的模板
var (
	templateCache = make(map[string]*template.Template)
	templateMutex sync.RWMutex
)

// GetTemplate 从缓存中获取模板，如果缓存中没有则加载并缓存
func GetTemplate(name string, funcMap template.FuncMap) (*template.Template, error) {
	// 先尝试从缓存中获取模板
	templateMutex.RLock()
	tmpl, ok := templateCache[name]
	templateMutex.RUnlock()
	if ok {
		return tmpl, nil
	}

	// 缓存中没有，加载模板
	templateMutex.Lock()
	defer templateMutex.Unlock()

	// 再次检查缓存，防止并发情况下重复加载
	if tmpl, ok := templateCache[name]; ok {
		return tmpl, nil
	}

	// 加载模板文件
	tmpl, err := template.New(name).Funcs(funcMap).ParseFS(TemplateFS, "templates/"+name)
	if err != nil {
		return nil, fmt.Errorf("找不到模板文件 %s: %w", name, err)
	}

	// 缓存模板
	templateCache[name] = tmpl
	return tmpl, nil
}

// RenderEnityTemplate
// 根据 entity模板渲染出一个Entity xxx.go
func RenderEnityTemplate(targetPath string, tableData *TableData) error {
	// 定义自定义函数
	funcMap := template.FuncMap{
		"unescaped": func(s string) template.HTML {
			return template.HTML(s)
		},
	}

	// 加载模板文件
	tmpl, err := GetTemplate("entity.tmpl", funcMap)
	if err != nil {
		return err
	}

	fileName := tableData.ObjectName + ".go"

	// 创建输出文件
	file, err := os.Create(filepath.Join(targetPath, fileName))
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败: %w", fileName, err)
	}

	defer file.Close()

	// 渲染模板并写入文件
	err = tmpl.Execute(file, tableData)
	if err != nil {
		return fmt.Errorf("渲染模板 %s 失败: %w", targetPath, err)
	}
	log.Println("Go file ", fileName, " generated successfully")
	return nil
}
