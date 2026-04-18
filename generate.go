package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"
)

type Document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
}

var baseTemplates = []string{
	"User logged into the system from IP %s",
	"Password reset requested by user %s",
	"Payment of $50.00 processed for invoice %s",
	"Error: Connection timeout reaching database shard %s",
	"New ticket created: Issue with login page for user %s",
	"User uploaded profile picture %s.png",
	"Background job %s completed successfully",
	"API Rate limit exceeded for client %s",
	"Monthly analytics report generated for tenant %s",
	"Customer feedback received from %s",
}

func main() {
	f, err := os.Create("100k_data.jsonl")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Use a fixed seed for reproducibility, or dynamic
	rand.Seed(time.Now().UnixNano())
	
	totalCount := 100000
	uniqueRatio := 0.02 // 2% unique, 98% duplicates
	uniqueCount := int(float64(totalCount) * uniqueRatio)

	fmt.Printf("Generating %d records with ~%d unique contents (98%% duplication)...\n", totalCount, uniqueCount)

	// Pre-generate the unique contents (2000 unique strings)
	uniqueContents := make([]string, uniqueCount)
	for i := 0; i < uniqueCount; i++ {
		template := baseTemplates[rand.Intn(len(baseTemplates))]
		uniqueContents[i] = fmt.Sprintf(template, fmt.Sprintf("ID_%d", i))
	}

	encoder := json.NewEncoder(f)
	
	for i := 0; i < totalCount; i++ {
		// Randomly select one of the 2000 unique strings
		content := uniqueContents[rand.Intn(uniqueCount)]
		
		doc := Document{
			ID:      fmt.Sprintf("doc_%d", i),
			Content: content,
			Metadata: map[string]string{
				"source": "archive",
				"time":   time.Now().Format(time.RFC3339),
			},
		}
		
		if err := encoder.Encode(doc); err != nil {
			panic(err)
		}
	}
	
	fmt.Println("✅ Done generating 100k_data.jsonl")
}
