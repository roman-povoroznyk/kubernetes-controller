package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Build information (set by ldflags)
var (
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "show version information",
	Long:  `Display version information including build details and Go version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("k6s version %s\n", Version)
		fmt.Printf("go version: %s\n", GoVersion)
		fmt.Printf("runtime: %s/%s\n", runtime.GOOS, runtime.GOARCH)
		fmt.Printf("build time: %s\n", BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
