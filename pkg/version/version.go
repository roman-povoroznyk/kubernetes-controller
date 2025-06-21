package version

// Build-time variables
var (
	Version   = "0.1.0"  // Will be overridden by ldflags
	GitCommit = "unknown" // Will be overridden by ldflags
	BuildDate = "unknown" // Will be overridden by ldflags
)

// GetVersion returns version information
func GetVersion() string {
	return Version
}

// GetBuildInfo returns complete build information
func GetBuildInfo() map[string]string {
	return map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_date": BuildDate,
	}
}
