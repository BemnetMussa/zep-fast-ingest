package lsh

import (
	"fmt"
	"sync"
)

const (
	ShardCount = 256 // We split our map into 256 shards to reduce lock contention
)

type Shard struct {
	sync.RWMutex
	buckets map[string]string
}

type Deduplicator struct {
	// Instead of one map, we have a slice of 256 independent shards
	shards [ShardCount]*Shard
}

func NewDeduplicator() *Deduplicator {
	d := &Deduplicator{}
	for i := 0; i < ShardCount; i++ {
		d.shards[i] = &Shard{
			buckets: make(map[string]string),
		}
	}
	return d
}

// getShard identifies which shard a band belongs to using its hash
func (d *Deduplicator) getShard(bandHash string) *Shard {
	// A simple checksum/hash of the bandHash to pick a shard 0-255
	var sum uint32
	for _, char := range bandHash {
		sum += uint32(char)
	}
	return d.shards[sum%ShardCount]
}

func (d *Deduplicator) IsDuplicate(docID string, signature []uint64) bool {
	isDup := false

	for bandIdx := 0; bandIdx < Bands; bandIdx++ {
		start := bandIdx * RowsPerBand
		end := start + RowsPerBand
		bandSlice := signature[start:end]
		bandHash := fmt.Sprintf("%d-%v", bandIdx, bandSlice) // Include bandIdx to prevent cross-band collisions

		// Get the specific shard for this band
		shard := d.getShard(bandHash)

		shard.Lock()
		if existingDocID, exists := shard.buckets[bandHash]; exists && existingDocID != docID {
			isDup = true
			shard.Unlock()
		} else {
			shard.buckets[bandHash] = docID
			shard.Unlock()
		}
	}

	return isDup
}
