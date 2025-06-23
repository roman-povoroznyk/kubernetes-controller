package main

import (
	"github.com/roman-povoroznyk/kubernetes-controller/cmd"

	// Import subpackages to register their commands
	_ "github.com/roman-povoroznyk/kubernetes-controller/cmd/kubernetes"
	_ "github.com/roman-povoroznyk/kubernetes-controller/cmd/server"
)

func main() {
	cmd.Execute()
}
