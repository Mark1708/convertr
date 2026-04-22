package main

import (
	"os"

	"github.com/Mark1708/convertr/internal/cli"
)

var Version = "dev"

func main() {
	if err := cli.New(Version).Execute(); err != nil {
		os.Exit(1)
	}
}
