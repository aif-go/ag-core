package cmd_project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"

	"github.com/aif-go/ag-core/tool/cmd/aggo/internal/base"
)

// Project is a project template.
type Project struct {
	Name string
	Path string
}

// New new a project from remote repo.
func (p *Project) New(ctx context.Context, dir string, layout string, branch string) error {
	to := filepath.Join(dir, p.Name)
	// Check if the folder already exists
	if _, err := os.Stat(to); !os.IsNotExist(err) {
		fmt.Printf("🚫 %s already exists\n", p.Name)
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		var override bool
		e := survey.AskOne(prompt, &override)
		if e != nil {
			return e
		}
		if !override {
			return err
		}
		err := os.RemoveAll(to)
		if err != nil {
			return err
		}
	}
	fmt.Printf("🚀 Creating service %s, layout repo is %s, please wait a moment.\n\n", p.Name, layout)
	repo := base.NewRepo(layout, branch)

	ignore := []string{
		".git",
		".github",
		".vscode",
	}

	replace := []string{}

	// if err := repo.CopyTo(ctx, to, p.Name, []string{".git", ".github"}); err != nil {
	if err := repo.CopyToV2(ctx, to, p.Name, ignore, replace); err != nil {
		return err
	}
	// e := os.Rename(
	// 	filepath.Join(to, "cmd", "server"),
	// 	filepath.Join(to, "cmd", p.Name),
	// )
	// if e != nil {
	// 	return e
	// }
	base.Tree(to, dir)

	fmt.Printf("\n🍺 Project creation succeeded %s\n", color.GreenString(p.Name))
	// fmt.Print("💻 Use the following command to start the project 👇:\n\n")
	// TODO
	// fmt.Println(color.WhiteString("TODO ...."))
	// fmt.Println(color.WhiteString("$ go generate ./..."))
	// fmt.Println(color.WhiteString("$ go build -o ./bin/ ./... "))
	// fmt.Println(color.WhiteString("$ ./bin/%s -conf ./configs\n", p.Name))
	fmt.Println("			🤝 Thanks for using Aggo")
	return nil
}
