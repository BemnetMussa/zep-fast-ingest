package lsh

import (
	"fmt"
	"sync"
)

// Deduplicator manages our local LSH buckets
type Deduplicator struct {
	// A map of BandIndex -> HashOfBand -> DocumentID
	// This acts as our localized in-memory database of seen text
	buckets []map[string]string
	mu sync.RWMutex
}

func NewDeduplicator() *Deduplicator {
	d := &Deduplicator{
		buckets: make([]map[string]string, Bands),
	}
	for i := 0; i < Bands; i++ {
		d.buckets[i] = make(map[string]string)
	}
	return d
}

// IsDuplicate checks the LSH bands. If ANY band matches a previous document,
// it flags it as a duplicate, saving an expensive Zep API call.
func (d *Deduplicator) IsDuplicate(docID string, signature[]uint64) bool {
	isDup := false
	d.mu.Lock()         // Lock for writing/reading
	defer d.mu.Unlock() // Unlock when done

	for bandIdx := 0; bandIdx < Bands; bandIdx++ {
		// Extract the 'RowsPerBand' (e.g., 5 numbers) for this band
		start := bandIdx * RowsPerBand
		end := start + RowsPerBand
		bandSlice := signature[start:end]

		// Create a string hash of this specific band
		bandHash := fmt.Sprintf("%v", bandSlice)

		// Check if we've seen this exact band hash before
		if existingDocID, exists := d.buckets[bandIdx][bandHash]; exists && existingDocID != docID {
			isDup = true
			// We don't break immediately because we want to finish adding 
			// the current document's bands to the buckets for future comparisons.
		} else {
			// Add this document's band to our bucket
			d.buckets[bandIdx][bandHash] = docID
		}
	}

	return isDup
}
