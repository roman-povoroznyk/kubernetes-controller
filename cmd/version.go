package cmd

import (
	"fmt"

	"github.com/roman-povoroznyk/k6s/pkg/logger"
	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of k6s",
	Long:  `Print the version number and build information for k6s.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Displaying version information", map[string]interface{}{
			"version": Version,
			"command": "version",
		})
		fmt.Printf("k6s version %s\n", Version)
		
		logger.Debug("Version command completed successfully", nil)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
