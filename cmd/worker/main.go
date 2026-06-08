package main

import (
	"fmt"
	"os"

	"github.com/juanlou1217/exitfleet/internal/config"
	"github.com/juanlou1217/exitfleet/internal/harness"
)

func main() {
	cfg := config.DefaultWorker()
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "worker configuration error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("%s worker proxy listening on %s via %s\n", harness.ProjectName, cfg.ProxyAddress(), cfg.TunDevice)
}

