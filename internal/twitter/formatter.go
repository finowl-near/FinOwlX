package twitter

import (
	"strings"
)

// splitCryptoTweet intelligently extracts each token-related segment from a large tweet string.
func SplitCryptoTweet(tweet string) []string {
	// Define the delimiter used in the AI prompt
	delimiter := "===PROJECT_BREAK==="

	// Split the tweet based on the delimiter
	parts := strings.Split(tweet, delimiter)

	// Trim whitespace and remove empty segments
	var result []string
	for _, part := range parts {
		cleaned := strings.TrimSpace(part)
		if cleaned != "" {
			result = append(result, cleaned)
		}
	}

	return result
}
