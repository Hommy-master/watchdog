// Test helper binary: runs until killed or for a fixed duration.
//
//go:build ignore

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	duration := flag.Int("duration", 0, "seconds to run before exit; 0 means wait for signal")
	flag.Parse()

	if *duration > 0 {
		time.Sleep(time.Duration(*duration) * time.Second)
		fmt.Println("exit")
		os.Exit(0)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh
}
