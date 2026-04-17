package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Document struct {
	ID      string `json:"id"`
	Content string `json:"content"`
}

func main() {
	file, _ := os.Create("data.jsonl")
	defer file.Close()

	fmt.Println("Generating 100,000 records...")

	for i := 0; i < 100000; i++ {
		doc := Document{
			ID:      fmt.Sprintf("msg_%d", i),
			Content: "This is a standard message for Zep AI context. User Bemnet is testing performance.",
		}

		// Every 5th message, we make it a slightly different duplicate
		if i%5 == 0 {
			doc.Content = "This is a standard message for Zep AI context. User Bemnet is testing performance!"
		}
		// Every 10th message, we make it unique
		if i%10 == 0 {
			doc.Content = fmt.Sprintf("Unique message number %d: Ethiopia is the land of origins.", i)
		}

		data, _ := json.Marshal(doc)
		file.Write(data)
		file.Write([]byte("\n"))
	}
	fmt.Println("Done! data.jsonl created.")
}
