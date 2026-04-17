package types

import "context"

// Document represents a single row of historical data (e.g., a chat message)
type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

// Streamer is responsible for reading large files line-by-line without blowing up RAM
type Streamer interface {
	// Stream pushes documents to a channel and closes it when done or on error
	Stream(ctx context.Context, filePath string, outChan chan<- Document) error
}

// Deduplicator is our MinHash/LSH engine.
type Deduplicator interface {
	// IsDuplicate checks if a document is highly similar to one we've already seen
	IsDuplicate(doc Document) (bool, error)
}

// ZepClient handles the actual network requests to Zep AI
type ZepClient interface {
	// SendBatch sends a slice of documents concurrently to Zep
	SendBatch(ctx context.Context, batch[]Document) error
}
