package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/bitrok/bitrok/server/internal/api"
	"github.com/bitrok/bitrok/server/internal/config"
)

func main() {
	var configPath string
	var showVersion bool
	flag.StringVar(&configPath, "config", "", "Path to config JSON file (optional)")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Println("bitrok-server", api.Version)
		os.Exit(0)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bitrok-server: failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := api.Run(cfg); err != nil {
		slog.Error("server exited", "error", err)
		os.Exit(1)
	}
}
