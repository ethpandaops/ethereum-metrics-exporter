package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ethpandaops/ethereum-metrics-exporter/cmd"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGTERM, syscall.SIGINT)

	done := make(chan error, 1)

	// Start application with context
	go func() {
		err := cmd.ExecuteWithContext(ctx)
		done <- err
	}()

	// Wait for either completion or signal
	select {
	case err := <-done:
		// Command completed (help, error, etc.)
		if err != nil {
			log.Printf("Application error: %v", err)
			os.Exit(1)
		}
		// Normal completion (like --help)
		return

	case sig := <-signalChan:
		// Signal received while running
		log.Printf("Caught signal: %v, initiating graceful shutdown...", sig)

		// Cancel context to trigger cleanup in all modules
		cancel()

		// Give modules time to cleanup (with timeout)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)

		// Wait for either shutdown completion or timeout
		select {
		case <-done:
			log.Println("Graceful shutdown complete")
		case <-shutdownCtx.Done():
			log.Println("Shutdown timeout reached, forcing exit")
		}

		shutdownCancel()
		os.Exit(0)
	}
}
