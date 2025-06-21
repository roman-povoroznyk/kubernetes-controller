package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestMainFunction(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "k6s_test", ".")
	cmd.Dir = "."
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("k6s_test")
	
	// Test version flag
	cmd = exec.Command("./k6s_test", "--version")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run version command: %v", err)
	}
	
	if len(output) == 0 {
		t.Error("Version output should not be empty")
	}
}

func TestCLICommands(t *testing.T) {
	// Build the binary
	cmd := exec.Command("go", "build", "-o", "k6s_test", ".")
	err := cmd.Run()
	if err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove("k6s_test")
	
	// Test help command
	cmd = exec.Command("./k6s_test", "--help")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run help command: %v", err)
	}
	
	if len(output) == 0 {
		t.Error("Help output should not be empty")
	}
	
	// Test server help
	cmd = exec.Command("./k6s_test", "server", "--help")
	output, err = cmd.Output()
	if err != nil {
		t.Fatalf("Failed to run server help command: %v", err)
	}
	
	if len(output) == 0 {
		t.Error("Server help output should not be empty")
	}
}
