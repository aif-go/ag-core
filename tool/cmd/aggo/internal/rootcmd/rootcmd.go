package rootcmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:     "aggo",
	Short:   "aggo: An elegant toolkit for Go microservices.",
	Long:    `aggo: An elegant toolkit for Go microservices.`,
	Version: "",
}

// SetVersion sets the version of the root command.
func SetVersion(version string) {
	rootCmd.Version = version
}

// RootCmd returns the root command.
func RootCmd() *cobra.Command {
	return rootCmd
}

// RegCommand register a command to root command, SPI for other command.
func RegCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

// Run executes the root command.
func Run() error {
	return rootCmd.Execute()
}
