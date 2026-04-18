// Package zepclient manages all network interactions with the Zep Cloud API.
// It translates our internal Pipeline documents into Zep Cloud's Temporal V2 entities.
package zepclient

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/bemnetmussa/zep-fast-ingest/internal/worker"
	"github.com/bemnetmussa/zep-fast-ingest/types"
)

// loadEnv parses a .env file natively so we don't need external dependencies
func loadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		return // Ignore if no .env file
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			os.Setenv(parts[0], strings.Trim(parts[1], `"'`))
		}
	}
}

// ProcessResults listens to a channel of deduplicated results and chunks them.
// It ensures that the required User and Thread exist on Zep Cloud before dispatching batches.
func ProcessResults(ctx context.Context, resultChan <-chan worker.Result, batchSize int, threadID string, userID string) {
	loadEnv()
	apiKey := os.Getenv("ZEP_API_KEY")
	apiURL := os.Getenv("ZEP_API_URL")

	if apiKey == "" || apiURL == "" {
		log.Fatal("❌ Missing ZEP_API_KEY or ZEP_API_URL environment variables. Please update your .env file.")
	}

	// 0. Ensure User Exists
	createUser(ctx, apiURL, apiKey, userID)

	// 1. Ensure Thread Exists (Zep V2 works with Threads and Messages)
	createThread(ctx, apiURL, apiKey, threadID)

	var batch []types.Document

	for {
		select {
		case <-ctx.Done():
			sendMessagesToZep(ctx, apiURL, apiKey, threadID, batch)
			return
		case res, ok := <-resultChan:
			if !ok {
				sendMessagesToZep(ctx, apiURL, apiKey, threadID, batch)
				return
			}

			if !res.IsDuplicate {
				batch = append(batch, res.Doc)
			}

			// Note: Zep Cloud enforces a max of 30 messages per POST request
			if len(batch) >= 30 {
				sendMessagesToZep(ctx, apiURL, apiKey, threadID, batch)
				batch = nil
			}
		}
	}
}

// createThread initializes a new thread context for the specified user within Zep Cloud.
func createThread(ctx context.Context, apiURL, apiKey, threadID string) {
	body, _ := json.Marshal(map[string]interface{}{
		"thread_id": threadID,
		"user_id":   "admin_uploader",
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v2/threads", apiURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	// API Key Header: Try "Api-Key" first.
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("⚠️  Could not reach Zep at %s: %v\n", apiURL, err)
		return
	}
	defer resp.Body.Close()

	errBody := make([]byte, 1024)
	n, _ := resp.Body.Read(errBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("🛠️  Thread '%s' successfully created on Zep Cloud.\n", threadID)
	} else if resp.StatusCode == 400 && (strings.Contains(string(errBody[:n]), "already exists") || strings.Contains(string(errBody[:n]), "already")) {
		fmt.Printf("🛠️  Thread '%s' verified (already exists).\n", threadID)
	} else {
		fmt.Printf("⚠️  Failed to create thread '%s'. Status: %d, Response: %s\n", threadID, resp.StatusCode, string(errBody[:n]))
	}
}

// createUser initializes a user in Zep Cloud. Threads must be attached to an existing user.
func createUser(ctx context.Context, apiURL, apiKey, userID string) {
	body, _ := json.Marshal(map[string]interface{}{
		"user_id": userID,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v2/users", apiURL), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("⚠️  Could not reach Zep at %s: %v\n", apiURL, err)
		return
	}
	defer resp.Body.Close()

	errBody := make([]byte, 1024)
	n, _ := resp.Body.Read(errBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("👤  User '%s' successfully created on Zep Cloud.\n", userID)
	} else if resp.StatusCode == 400 && (strings.Contains(string(errBody[:n]), "already exists") || strings.Contains(string(errBody[:n]), "already")) {
		fmt.Printf("👤  User '%s' verified (already exists).\n", userID)
	} else {
		fmt.Printf("⚠️  Failed to create user '%s'. Status: %d, Response: %s\n", userID, resp.StatusCode, string(errBody[:n]))
	}
}

// sendMessagesToZep formats internal documents into Zep Messages and POSTs them to the specified Thread.
func sendMessagesToZep(ctx context.Context, apiURL, apiKey, threadID string, batch []types.Document) {
	if len(batch) == 0 {
		return
	}

	fmt.Printf("📡 [Zep Client] Sending batch of %d unique messages to Zep Cloud...\n", len(batch))

	type zepMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	var msgs []zepMessage
	for _, d := range batch {
		msgs = append(msgs, zepMessage{
			Role:    "user",
			Content: d.Content,
		})
	}

	body, _ := json.Marshal(map[string]interface{}{
		"messages": msgs,
	})

	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("%s/api/v2/threads/%s/messages", apiURL, threadID),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Api-Key %s", apiKey))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send batch to Zep: %v", err)
		return
	}
	defer resp.Body.Close()

	errBody := make([]byte, 1024)
	n, _ := resp.Body.Read(errBody)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Printf("✅ [Zep API] Successfully ingested %d unique messages. (Status: %d)\n", len(batch), resp.StatusCode)
	} else {
		fmt.Printf("❌ [Zep API] Failed to ingest batch. Status: %d, Response: %s\n", resp.StatusCode, string(errBody[:n]))
		
		// Fallback diagnostic hint
		if resp.StatusCode == 401 && strings.Contains(string(errBody[:n]), "Unauthorized") {
			fmt.Println("👉 Tip: If 'Api-Key XXX' is failing, try changing req.Header.Set to 'Bearer XXX' locally.")
		}
	}
}
