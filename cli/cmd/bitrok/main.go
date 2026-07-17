package main

import (
	"fmt"
	"os"

	"github.com/bitrok/bitrok/cli/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "bitrok: %v\n", err)
		os.Exit(1)
	}
}
