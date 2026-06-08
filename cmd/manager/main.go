package main

import (
	"fmt"
	"os"

	"github.com/juanlou1217/exitfleet/internal/config"
	"github.com/juanlou1217/exitfleet/internal/harness"
)

func main() {
	cfg := config.DefaultManager()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "manager configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s manager listening on %s\n", harness.ProjectName, cfg.HTTPAddress())
}

