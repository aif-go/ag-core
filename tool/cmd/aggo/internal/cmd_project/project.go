package cmd_project

import (
	"ag-core/tool/cmd/aggo/internal/rootcmd"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var (
	repoURL string
	branch  string
	timeout string
	nomod   bool
)

func init() {
	rootcmd.RegCommand(CmdNew)

	if repoURL = os.Getenv("AG_LAYOUT_REPO"); repoURL == "" {
		// repoURL = "https://github.com/frochyzhang/ag-layout.git"
		repoURL = "https://github.com/ZhengweiHou/ag-layout-demo.git"
	}
	timeout = "120s"

	CmdNew.Flags().StringVarP(&repoURL, "repo-url", "r", repoURL, "layout repo")
	CmdNew.Flags().StringVarP(&branch, "branch", "b", branch, "repo branch")
	CmdNew.Flags().StringVarP(&timeout, "timeout", "t", timeout, "time out")
	CmdNew.Flags().BoolVarP(&nomod, "nomod", "", nomod, "retain go mod")
}

// CmdNew represents the new command.
var CmdNew = &cobra.Command{
	Use:   "new",
	Short: "Create a service template",
	Long:  "Create a service project using the repository template. Example: aggo new helloworld",
	Run:   run,
}

func run(_ *cobra.Command, args []string) {

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	t, err := time.ParseDuration(timeout)
	if err != nil {
		panic(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), t)
	defer cancel()

	name := ""
	if len(args) == 0 {
		prompt := &survey.Input{
			Message: "What is project name ?",
			Help:    "Created project name.",
		}
		err = survey.AskOne(prompt, &name)
		if err != nil || name == "" {
			return
		}
	} else {
		name = args[0]
	}

	projectName, workingDir := processProjectParams(name, wd)
	p := &Project{Name: projectName}
	done := make(chan error, 1)
	go func() {
		if !nomod {
			done <- p.New(ctx, workingDir, repoURL, branch)
			return
		}
	}()
	select {
	case <-ctx.Done():
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Fprint(os.Stderr, "\033[31mERROR: project creation timed out\033[m\n")
			return
		}
		fmt.Fprintf(os.Stderr, "\033[31mERROR: failed to create project(%s)\033[m\n", ctx.Err().Error())
	case err = <-done:
		if err != nil {
			fmt.Fprintf(os.Stderr, "\033[31mERROR: Failed to create project(%s)\033[m\n", err.Error())
		}
	}
}

// processProjectParams 处理项目名称和工作目录。
// 它将项目名称中的“~”前缀扩展为用户的主目录，
// 并确保项目路径是绝对路径。
// 它返回项目的基本名称和项目所在的目录。
func processProjectParams(projectName string, workingDir string) (projectNameResult, workingDirResult string) {
	_projectDir := projectName
	_workingDir := workingDir
	// Process ProjectName with system variable
	if strings.HasPrefix(projectName, "~") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			// cannot get user home return fallback place dir
			return _projectDir, _workingDir
		}
		_projectDir = filepath.Join(homeDir, projectName[2:])
	}

	// check path is relative
	if !filepath.IsAbs(projectName) {
		absPath, err := filepath.Abs(projectName)
		if err != nil {
			return _projectDir, _workingDir
		}
		_projectDir = absPath
	}

	return filepath.Base(_projectDir), filepath.Dir(_projectDir)
}
