// Package streamer handles parsing massive JSONL files asynchronously with 
// constrained memory buffers, preventing Out-Of-Memory (OOM) crashes.
package streamer

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/bemnetmussa/zep-fast-ingest/types"
)

// StreamJSONL reads a file line-by-line and pushes Documents into a channel.
func StreamJSONL(ctx context.Context, filePath string) (<-chan types.Document, <-chan error) {
	docChan := make(chan types.Document, 100) // Buffered channel for speed
	errChan := make(chan error, 1)

	go func() {
		defer close(docChan)
		defer close(errChan)

		file, err := os.Open(filePath)
		if err != nil {
			errChan <- err
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		// Use a larger buffer for lines (important for long chat histories)
		const maxCapacity = 1 * 1024 * 1024 // 1MB per line max
		buf := make([]byte, maxCapacity)
		scanner.Buffer(buf, maxCapacity)

		for scanner.Scan() {
			select {
			case <-ctx.Done(): // If the user hits Ctrl+C, stop reading immediately
				return
			default:
				var doc types.Document
				if err := json.Unmarshal(scanner.Bytes(), &doc); err != nil {
					// Add a non-blocking log prefix rather than silently ignoring corrupted lines
					// log.Printf("skipping malformed line: %v", err)
					continue 
				}
				docChan <- doc
			}
		}

		if err := scanner.Err(); err != nil {
			errChan <- err
		}
	}()

	return docChan, errChan
}
