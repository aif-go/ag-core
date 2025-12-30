package cmd_proto

import (
	"ag-core/tool/cmd/aggo/internal/rootcmd"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var (
	desc       bool
	clientMode bool
	serverMode bool
	allMode    bool

	extProtoPath []string
	plugins      []string
	models       []string

	extPlugins []string
)

func init() {
	// 注册插件
	initRegPlugins()

	rootcmd.RegCommand(CmdProto)

	CmdProto.Flags().BoolVarP(&desc, "desc", "d", desc, "show debug log")

	// CmdProto.Flags().BoolVarP(&clientMode, "client", "c", clientMode, "generate client code")
	// CmdProto.Flags().BoolVarP(&serverMode, "server", "s", serverMode, "generate server code")
	// CmdProto.Flags().BoolVarP(&allMode, "all", "a", allMode, "generate all code")

	CmdProto.Flags().StringArrayVarP(&extProtoPath, "ext-proto-path", "e", []string{}, "external proto path, eg: -e ./idl")
	CmdProto.Flags().StringArrayVarP(&plugins, "plugins", "p", []string{"all"}, fmt.Sprintf("plugins name, eg: -p all,%s", strings.Join(GetAllPluginsName(), ",")))
	CmdProto.Flags().StringArrayVarP(&models, "models", "m", []string{"all"}, "models name, eg: -m all,server,client")
	CmdProto.Flags().StringArrayVarP(&extPlugins, "ext-plugins", "P", []string{}, "extra plugins name, default none, eg: -P '--xxx'")
}

// CmdProto represents the proto command.
var CmdProto = &cobra.Command{
	Use:   "proto [flags] <idl-file-or-directory>...",
	Short: "Generate the proto files",
	Long:  "Generate the proto files.",
	Run:   run,
}

var (
	l_plugins    []string
	l_extPlugins []string
	l_protos     []string
	l_models     []string
)

func run(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		fmt.Println("Please enter the proto file or directory")
		return
	}

	err := initProtos(args)
	if err != nil {
		slog.Error("init protos failed", "err", err)
		os.Exit(1)
	}

	err = initPlugins()
	if err != nil {
		slog.Error("init plugins failed", "err", err)
		os.Exit(1)
	}

	err = initExtPlugins()
	if err != nil {
		slog.Error("init ext plugins failed", "err", err)
		os.Exit(1)
	}

	err = initModels()
	if err != nil {
		slog.Error("init models failed", "err", err)
		os.Exit(1)
	}

	err = generate()
	if err != nil {
		slog.Error("generate failed", "err", err)
		os.Exit(1)
	}
}

func generate() error {

	slog.Info("plugins", "plugins", l_plugins)
	slog.Info("ext plugins", "extPlugins", extPlugins)
	slog.Info("models", "models", l_models)
	slog.Info("proto", "protos", l_protos)

	input := []string{
		// "--proto_path=.",
	}

	// add ext proto path
	for _, p := range extProtoPath {
		if pathExists(p) {
			input = append(input, "--proto_path="+p)
		} else {
			err := fmt.Errorf("ext proto path not exists: %s", p)
			slog.Error("init ext proto path failed", "err", err)
			return err
		}
	}

	if pathExists("./third_party") {
		input = append(input, "--proto_path=./third_party")
	}

	inputExt, err := selectPlugins(l_plugins, l_models)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// inputExt := []string{
	// 	"--go_out=paths=source_relative:.",
	// 	"--go-agserver_out=model=all:.",
	// 	"--go-aghertz_out=model=client:.",
	// 	"--go-aghertz_out=model=server:.",
	// 	"--go-agkitex_out=model=server:.",
	// 	"--go-agkitex_out=model=client:.",
	// 	"--go-agservice_out=.",
	// }
	input = append(input, inputExt...)
	// add ext plugins
	input = append(input, l_extPlugins...)

	input = append(input, l_protos...)

	input = lo.Uniq(input) // 去重

	commandstr := "protoc " + strings.Join(input, " ")

	fmt.Printf("command: %s\n\n", strings.ReplaceAll(commandstr, " ", "\n "))
	if !desc {
		fd := exec.Command("protoc", input...)
		fd.Stdout = os.Stdout
		fd.Stderr = os.Stderr
		fd.Dir = "."
		if err := fd.Run(); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func initModels() error {
	if allMode {
		l_models = append(l_models, ModelAll)
	}
	if clientMode {
		l_models = append(l_models, ModelClient)
	}
	if serverMode {
		l_models = append(l_models, ModelServer)
	}

	for _, m := range models {
		if strings.Contains(m, ",") {
			l_models = append(l_models, strings.Split(m, ",")...)
		} else {
			l_models = append(l_models, m)
		}
	}

	l_models = lo.Uniq(l_models)

	return nil
}

func initPlugins() error {
	for _, p := range plugins {
		if p == "" {
			continue
		}
		if strings.Contains(p, ",") {
			l_plugins = append(l_plugins, strings.Split(p, ",")...)
		} else {
			l_plugins = append(l_plugins, p)
		}
	}
	l_plugins = lo.Uniq(l_plugins)
	return nil
}

func initExtPlugins() error {
	for _, p := range extPlugins {
		if p == "" {
			continue
		}

		// TODO 防非法注入检查

		l_extPlugins = append(l_extPlugins, p)
	}
	l_extPlugins = lo.Uniq(l_extPlugins)
	return nil
}

func initProtos(args []string) error {
	var err error
	l_protos, err = findProtos(args)

	return err
}
