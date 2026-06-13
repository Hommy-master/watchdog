package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"watchdog/internal/config"
	"watchdog/internal/logger"
	"watchdog/internal/monitor"
)

func main() {
	if err := logger.Init(); err != nil {
		panic(err)
	}
	defer logger.Close()

	configPath := flag.String("config", "config.json", "path to JSON config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatalf("load config: %v", err)
	}

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
