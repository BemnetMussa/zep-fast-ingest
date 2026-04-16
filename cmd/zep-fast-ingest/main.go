package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 1. Setup graceful shutdown context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = ctx
	// Listen for Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n[!] Shutting down gracefully... allowing current batches to finish.")
		cancel() // This cancels the context, telling all goroutines to wrap up
	}()

	fmt.Println("🚀 Starting Zep-Fast-Ingest Pipeline...")
	start := time.Now()

	// TODO: Initialize Streamer
	// TODO: Initialize LSH Deduplicator
	// TODO: Initialize Zep Client
	// TODO: Start Worker Pool

	// Simulate work for now
	time.Sleep(2 * time.Second)

	fmt.Printf("✅ Ingestion completed in %v\n", time.Since(start))
}
