package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"watchdog/internal/config"
	"watchdog/internal/logger"
	"watchdog/internal/monitor"
)

func main() {
	configPath := flag.String("config", "config.json", "path to JSON config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	mon := monitor.New(cfg)
	logger.Printf("watchdog started: interval=%ds delay=%ds apps=%d", cfg.Interval, cfg.Delay, len(cfg.Apps))
	if err := mon.Run(ctx); err != nil && err != context.Canceled {
		logger.Fatalf("monitor: %v", err)
	}
	logger.Println("watchdog stopped")
}
