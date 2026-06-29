package main

import (
	"os"

	"github.com/charmbracelet/log"
	"github.com/nkootstra/cli/internal/cli"
)

func main() {
	if err := cli.NewRootCommand().Execute(); err != nil {
		log.Error("command failed", "err", err)
		os.Exit(1)
	}
}
