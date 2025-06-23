package server

import (
	"testing"
)

func TestServerCommandDefined(t *testing.T) {
	if serverCmd == nil {
		t.Fatal("serverCmd should be defined")
	}
	if serverCmd.Use != "server" {
		t.Errorf("expected command use 'server', got %s", serverCmd.Use)
	}

	// Перевіряємо наявність прапора в самій команді
	portFlag := serverCmd.Flags().Lookup("server-port")
	if portFlag == nil {
		t.Error("expected 'server-port' flag to be defined")
	}

	// Перевіряємо значення за замовчуванням
	defaultValue := portFlag.DefValue
	if defaultValue != "8080" {
		t.Errorf("expected default port 8080, got %s", defaultValue)
	}
}
