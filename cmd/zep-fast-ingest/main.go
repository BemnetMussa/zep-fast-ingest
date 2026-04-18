// Package main is the entry point for the Zep-Fast-Ingest pipeline.
// It orchestrates the streamer, worker pool, and Zep Cloud batcher.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bemnetmussa/zep-fast-ingest/internal/lsh"
	"github.com/bemnetmussa/zep-fast-ingest/internal/streamer"
	"github.com/bemnetmussa/zep-fast-ingest/internal/worker"
	"github.com/bemnetmussa/zep-fast-ingest/internal/zepclient"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Setup Signal Handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("\n[!] Shutting down gracefully... allowing current batches to finish.")
		cancel()
	}()

	var filename, threadID, userID string
	flag.StringVar(&filename, "file", "data.jsonl", "JSONL file to ingest")
	flag.StringVar(&threadID, "thread", "fast_ingest_test_thread", "Zep Cloud Thread ID to ingest into")
	flag.StringVar(&userID, "user", "admin_uploader", "Zep Cloud User ID to associate with the thread")
	flag.Parse()

	fmt.Println("🚀 Starting Zep-Fast-Ingest Pipeline...")
	start := time.Now()

	// 2. Initialize Brain
	deduper := lsh.NewDeduplicator()

	// 3. Start Streamer
	docChan, errChan := streamer.StreamJSONL(ctx, filename)

	// 4. Start Worker Pool (10 parallel workers)
	resultChan := worker.StartWorkerPool(ctx, 10, docChan, deduper)

	// 5. THE CRITICAL STEP: Hand results to the Zep Batcher
	// This function will block until resultChan is closed and all batches are sent
	fmt.Println("📦 Processing stream and sending unique batches to Zep...")
	zepclient.ProcessResults(ctx, resultChan, 100, threadID, userID) // Initial batch buffer size of 100

	// 6. Final check for errors
	select {
	case err := <-errChan:
		if err != nil {
			fmt.Printf("❌ Streamer error: %v\n", err)
		}
	default:
	}

	fmt.Printf("\n✅ Pipeline Finished! Total Time: %v\n", time.Since(start))
}
