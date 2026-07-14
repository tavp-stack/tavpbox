package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tavpbox",
	Short: "Local dev environment — like Lando, but with Podman",
	Long: `TAVPBox — Lando Dockerless.
Local development environment using Podman containers.
One project = one container. Simple, fast, production-like.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	rootCmd.AddCommand(
		versionCmd,
		initCmd,
		createCmd,
		startCmd,
		stopCmd,
		restartCmd,
		destroyCmd,
		rebuildCmd,
		sshCmd,
		listCmd,
		infoCmd,
		logsCmd,
		toolingListCmd,
	)
	// Register dynamic tooling commands (artisan, composer, npm, etc.)
	RegisterToolingCommands()
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n  ✗ Error: %s\n\n", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tavpbox %s\n", Version)
	},
}
