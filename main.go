package main

import (
	"os"

	"github.com/dadoonet/soundbeat/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
