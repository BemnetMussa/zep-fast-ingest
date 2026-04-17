package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func main() {
	http.HandleFunc("/api/v1/collection/", func(w http.ResponseWriter, r *http.Request) {
		// Distinguish between /api/v1/collection/{name} and /api/v1/collection/{name}/document
		if strings.HasSuffix(r.URL.Path, "/document") {
			if r.Method == http.MethodPost {
				var body map[string]interface{}
				err := json.NewDecoder(r.Body).Decode(&body)
				if err != nil {
					http.Error(w, "invalid json", http.StatusBadRequest)
					return
				}
				docs := body["documents"].([]interface{})
				
				fmt.Printf("📡 [Mock] Received Batch: %d unique documents | Timestamp: %s\n", len(docs), time.Now().Format("15:04:05"))
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"message": "success"}`))
				return
			}
		}

		// Fallback for Create Collection
		if r.Method == http.MethodPost {
			fmt.Printf("🛠️  [Mock] Collection created/verified at: %s\n", r.URL.Path)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message": "success"}`))
			return
		}
	})

	fmt.Println("🛰️  Zep Mock Server running on http://localhost:8000")
	fmt.Println("   Press Ctrl+C to stop.")
	http.ListenAndServe(":8000", nil)
}
