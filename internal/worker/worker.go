// Package worker implements a producer-consumer concurrency model 
// to maximize CPU utilization during the expensive hashing computations.
package worker

import (
	"context"
	"sync"

	"github.com/bemnetmussa/zep-fast-ingest/internal/lsh"
	"github.com/bemnetmussa/zep-fast-ingest/types"
)

// Result holds the outcome of a worker's processing
type Result struct {
	Doc         types.Document
	IsDuplicate bool
}

// StartWorkerPool spawns configurable concurrent worker goroutines.
// It bridges the file streamer (Producer) and the Zep network batcher (Consumer).
func StartWorkerPool(ctx context.Context, numWorkers int, docChan <-chan types.Document, deduper *lsh.Deduplicator) <-chan Result {
	resultChan := make(chan Result, numWorkers*2)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case doc, ok := <-docChan:
					if !ok {
						return
					}

					// 1. Calculate Signature (Expensive CPU work)
					shingles := lsh.Shingle(doc.Content, 3)
					sig := lsh.GenerateSignature(shingles)

					// 2. Check Deduplicator (Thread-safe memory work)
					isDup := deduper.IsDuplicate(doc.ID, sig)

					// 3. Push Result
					resultChan <- Result{
						Doc:         doc,
						IsDuplicate: isDup,
					}
				}
			}
		}(i)
	}

	// Closer: Wait for all workers to finish, then close the results channel
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	return resultChan
}
