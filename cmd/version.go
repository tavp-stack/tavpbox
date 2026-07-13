package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show TAVPBox version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tavpbox %s\n", Version)
	},
}
