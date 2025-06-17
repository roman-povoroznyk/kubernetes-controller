package cmd

import (
"fmt"

"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of k6s",
	Long:  `Print the version number and build information for k6s.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("k6s version %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
