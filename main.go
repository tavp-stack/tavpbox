package main

import (
	"os"

	"github.com/tavp-stack/tavpbox/cmd"
)

var version = "dev"

func main() {
	cmd.Version = version
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
