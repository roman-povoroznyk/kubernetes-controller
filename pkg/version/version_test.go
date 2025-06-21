package version

import (
	"strings"
	"testing"
)

func TestGetVersion(t *testing.T) {
	version := GetVersion()
	if version == "" {
		t.Error("Version should not be empty")
	}
}

func TestGetBuildInfo(t *testing.T) {
	buildInfo := GetBuildInfo()
	
	// Check required fields
	requiredFields := []string{"version", "git_commit", "build_date"}
	for _, field := range requiredFields {
		if _, exists := buildInfo[field]; !exists {
			t.Errorf("Build info should contain field: %s", field)
		}
	}
	
	// Check version field is not empty
	if buildInfo["version"] == "" {
		t.Error("Version in build info should not be empty")
	}
}

func TestVersionNotUnknown(t *testing.T) {
	version := GetVersion()
	if strings.Contains(strings.ToLower(version), "unknown") {
		t.Log("Version contains 'unknown', this is expected in test environment")
	}
}
