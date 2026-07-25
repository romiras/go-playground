package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"play.ground/ollama-bench/internal/config"
	"play.ground/ollama-bench/internal/reporter"
	"play.ground/ollama-bench/internal/runner"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to YAML config")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("benchmarking model=%s hosts=%d timeout=%ds cleanup=%v",
		cfg.Model, len(cfg.Hosts), cfg.TimeoutSec, cfg.Cleanup)

	pool := runner.New(cfg)
	results := pool.Run(ctx)

	if err := reporter.WriteTable(os.Stdout, results); err != nil {
		log.Fatalf("table: %v", err)
	}

	if err := reporter.WriteJSON(cfg.Output, cfg.Model, results); err != nil {
		log.Fatalf("json export: %v", err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", cfg.Output)

	for _, r := range results {
		if r.Status == reporter.StatusFailed || r.Status == reporter.StatusUnreachable {
			os.Exit(1)
		}
	}
}
