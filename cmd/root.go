package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tavpbox",
	Short: "LXC-based dev environment — like Lando, but lighter",
	Long: `TAVPBox — dev environment all-in-one.

LXC-based system containers. ~30MB RAM per project.
Works on Linux (native), macOS (Lima), Windows (WSL2).`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(
		versionCmd,
		initCmd,
		createCmd,
		startCmd,
		stopCmd,
		listCmd,
		destroyCmd,
		rebuildCmd,
		sshCmd,
		snapshotCmd,
		infoCmd,
		statusCmd,
		logsCmd,
		execCmd,
	)
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
