package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of k8s-ctrl",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("k8s-ctrl version %s\n", appVersion)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
