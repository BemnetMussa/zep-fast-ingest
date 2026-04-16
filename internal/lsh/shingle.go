package lsh

import (
	"strings"
	"regexp"
)

var (
	// Pre-compile regex to remove punctuation for faster processing
	cleanRegex = regexp.MustCompile(`[^\w\s]`)
)

// Shingle breaks text into overlapping word n-grams.
// We return a map to automatically deduplicate shingles within the same document.
func Shingle(text string, n int) map[string]struct{} {
	// 1. Clean and normalize the text (lowercase, remove punctuation)
	text = strings.ToLower(text)
	text = cleanRegex.ReplaceAllString(text, "")
	
	// 2. Split into words
	words := strings.Fields(text)
	
	shingles := make(map[string]struct{})
	
	// 3. If the text is shorter than our n-gram size, just return it as one shingle
	if len(words) < n {
		if len(words) > 0 {
			shingles[strings.Join(words, " ")] = struct{}{}
		}
		return shingles
	}

	// 4. Slide a window over the words to create overlapping chunks
	for i := 0; i <= len(words)-n; i++ {
		chunk := strings.Join(words[i:i+n], " ")
		shingles[chunk] = struct{}{}
	}

	return shingles
}
