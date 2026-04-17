package zepclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/bemnetmussa/zep-fast-ingest/internal/worker"
	"github.com/bemnetmussa/zep-fast-ingest/types"
)

const baseURL = "http://localhost:8000"

func ProcessResults(ctx context.Context, resultChan <-chan worker.Result, batchSize int) {
	collectionName := "historic_chats"

	// 1. Ensure Collection Exists
	createCollection(ctx, collectionName)

	var batch []types.Document

	for {
		select {
		case <-ctx.Done():
			sendToZep(ctx, collectionName, batch)
			return
		case res, ok := <-resultChan:
			if !ok {
				sendToZep(ctx, collectionName, batch)
				return
			}

			if !res.IsDuplicate {
				batch = append(batch, res.Doc)
			}

			if len(batch) >= batchSize {
				sendToZep(ctx, collectionName, batch)
				batch = nil
			}
		}
	}
}

func createCollection(ctx context.Context, name string) {
	body, _ := json.Marshal(map[string]string{
		"name":        name,
		"description": "Bemnet's Fast Ingest Test",
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/collection/%s", baseURL, name), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("⚠️  Could not reach Zep at %s (is the mock server running?): %v\n", baseURL, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("🛠️  Collection '%s' verified/created.\n", name)
}

func sendToZep(ctx context.Context, collectionName string, batch []types.Document) {
	if len(batch) == 0 {
		return
	}

	fmt.Printf("📡 [Zep Client] Sending batch of %d unique documents to Mock API...\n", len(batch))

	// Convert to the format the mock server expects
	type zepDoc struct {
		DocumentID string                 `json:"document_id"`
		Content    string                 `json:"content"`
		Metadata   map[string]interface{} `json:"metadata"`
	}

	var docs []zepDoc
	for _, d := range batch {
		docs = append(docs, zepDoc{
			DocumentID: d.ID,
			Content:    d.Content,
			Metadata:   map[string]interface{}{"source": "fast-ingest-stress-test"},
		})
	}

	body, _ := json.Marshal(map[string]interface{}{
		"documents": docs,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v1/collection/%s/document", baseURL, collectionName),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send batch to Zep: %v", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("✅ [Zep API] Successfully ingested %d unique documents. (Status: %d)\n", len(batch), resp.StatusCode)
}
