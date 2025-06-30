package main

import (
	"os"

	"github.com/roman-povoroznyk/kubernetes-controller/k6s/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
