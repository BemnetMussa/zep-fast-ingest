package lsh

import (
	"hash/fnv"
	"math"
)

// Signature size (e.g., 100) and Bands are crucial for LSH.
// Zep relies on banding to find near-duplicates fast.
const (
	SignatureSize = 100
	Bands         = 20
	RowsPerBand   = SignatureSize / Bands // 5 rows per band
)

// GenerateSignature mimics the fast permutation MinHash approach.
func GenerateSignature(shingles map[string]struct{})[]uint64 {
	signature := make([]uint64, SignatureSize)
	for i := range signature {
		signature[i] = math.MaxUint64
	}

	if len(shingles) == 0 {
		return signature
	}

	for shingle := range shingles {
		// Use FNV-1a (Go's fast non-cryptographic hash)
		h := fnv.New64a()
		h.Write([]byte(shingle))
		baseHash := h.Sum64()

		// Simulate 'SignatureSize' different hash functions via linear permutation
		// Exact math trick used in Python data-sci libraries (like datasketc)
		for i := 0; i < SignatureSize; i++ {
			// (a * x + b) % c permutation approximation
			mixedHash := (baseHash ^ uint64(i)) * 1099511628211 

			if mixedHash < signature[i] {
				signature[i] = mixedHash
			}
		}
	}

	return signature
}
