package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var Version = "dev"

var rootCmd = &cobra.Command{
	Use:   "tavpbox",
	Short: "LXC-based dev environment — like Lando, but lighter",
	Long: `TAVPBox — dev environment all-in-one.

LXC-based system containers. ~30MB RAM per project.
Works on Linux (native), macOS (Lima), Windows (WSL2).`,
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
		listCmd,
		statusCmd,
		infoCmd,
		sshCmd,
		execCmd,
		logsCmd,
		snapshotCmd,
	)
}

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		printFriendlyError(err)
	}
	return err
}

func printFriendlyError(err error) {
	msg := err.Error()

	fmt.Fprintf(os.Stderr, "\n  ✗ Error: %s\n\n", msg)

	// Provide troubleshooting hints
	if strings.Contains(msg, "not found") || strings.Contains(msg, "not available") {
		fmt.Fprintln(os.Stderr, "  Troubleshooting:")
		fmt.Fprintln(os.Stderr, "    1. Run 'tavpbox setup' to install dependencies")
		fmt.Fprintln(os.Stderr, "    2. Make sure LXD is installed and running")
		fmt.Fprintln(os.Stderr, "    3. On Windows, ensure WSL2 is enabled")
		fmt.Fprintln(os.Stderr, "")
	}

	if strings.Contains(msg, "permission denied") || strings.Contains(msg, "access denied") {
		fmt.Fprintln(os.Stderr, "  Troubleshooting:")
		fmt.Fprintln(os.Stderr, "    1. Run as Administrator/root")
		fmt.Fprintln(os.Stderr, "    2. Check file permissions")
		fmt.Fprintln(os.Stderr, "")
	}

	if strings.Contains(msg, "timeout") {
		fmt.Fprintln(os.Stderr, "  Troubleshooting:")
		fmt.Fprintln(os.Stderr, "    1. Check your internet connection")
		fmt.Fprintln(os.Stderr, "    2. Try again later")
		fmt.Fprintln(os.Stderr, "    3. The server might be busy")
		fmt.Fprintln(os.Stderr, "")
	}

	if strings.Contains(msg, "already exists") {
		fmt.Fprintln(os.Stderr, "  Troubleshooting:")
		fmt.Fprintln(os.Stderr, "    1. Use 'tavpbox destroy <name>' to remove existing box")
		fmt.Fprintln(os.Stderr, "    2. Or use a different name")
		fmt.Fprintln(os.Stderr, "")
	}
}

func checkErr(err error) {
	if err != nil {
		printFriendlyError(err)
		os.Exit(1)
	}
}
