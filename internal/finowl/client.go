package finowl

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	BaseURL = "https://finowl.finance/api/v0/summary"
)

// Summary represents the structure of the Finowl API response
type Summary struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
}

// Response represents the full API response
type Response struct {
	Summary Summary `json:"summary"`
	Total   int     `json:"total"`
}

// ContentSections contains the parsed sections from the summary content
type ContentSections struct {
	FeaturedTickers    string
	InfluencerInsights string
	MarketSentiment    string
}

// Client handles interactions with the Finowl API
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Finowl API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		baseURL: BaseURL,
	}
}

// GetSummary fetches a summary by ID
func (c *Client) GetSummary(id int) (*Response, error) {
	url := fmt.Sprintf("%s?id=%d", c.baseURL, id)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, ErrAPIRequestFailed{Cause: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrSummaryNotFound{ID: id}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnexpectedStatusCode{StatusCode: resp.StatusCode}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, ErrAPIRequestFailed{Cause: fmt.Errorf("failed to read response body: %w", err)}
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, ErrAPIRequestFailed{Cause: fmt.Errorf("failed to unmarshal response: %w", err)}
	}

	return &response, nil
}

// ParseContent extracts the three main sections from the summary content
func (c *Client) ParseContent(content string) (*ContentSections, error) {
	// Define section headers
	featuredHeader := "## Featured Tickers and Projects"
	insightsHeader := "## Key Insights from Influencers"
	sentimentHeader := "## Market Sentiment and Directions"

	// Find the indices of each section
	featuredIndex := strings.Index(content, featuredHeader)
	insightsIndex := strings.Index(content, insightsHeader)
	sentimentIndex := strings.Index(content, sentimentHeader)

	// Check which sections are missing
	var missingSections []string
	if featuredIndex == -1 {
		missingSections = append(missingSections, "Featured Tickers and Projects")
	}
	if insightsIndex == -1 {
		missingSections = append(missingSections, "Key Insights from Influencers")
	}
	if sentimentIndex == -1 {
		missingSections = append(missingSections, "Market Sentiment and Directions")
	}

	if len(missingSections) > 0 {
		return nil, ErrMissingSections{MissingSections: missingSections}
	}

	// Extract each section
	featuredContent := content[featuredIndex+len(featuredHeader) : insightsIndex]
	insightsContent := content[insightsIndex+len(insightsHeader) : sentimentIndex]
	sentimentContent := content[sentimentIndex+len(sentimentHeader):]

	// Clean up the content for Twitter (remove markdown formatting, limit length)
	featuredContent = cleanForTwitter(featuredContent)
	insightsContent = cleanForTwitter(insightsContent)
	sentimentContent = cleanForTwitter(sentimentContent)

	return &ContentSections{
		FeaturedTickers:    featuredContent,
		InfluencerInsights: insightsContent,
		MarketSentiment:    sentimentContent,
	}, nil
}

// WaitForNextSummary polls until the next summary ID is available
func (c *Client) WaitForNextSummary(currentID int) (*Response, error) {
	nextID := currentID + 1

	for {
		summary, err := c.GetSummary(nextID)
		if err == nil {
			return summary, nil
		}

		// If it's a 404, wait and try again
		if _, ok := err.(ErrSummaryNotFound); ok {
			fmt.Printf("Summary ID %d not yet available, waiting 15 minutes...\n", nextID)
			time.Sleep(15 * time.Minute)
			continue
		}

		// For other errors, return immediately
		return nil, err
	}
}

// cleanForTwitter prepares content for Twitter by removing markdown and limiting length
func cleanForTwitter(content string) string {
	// Remove markdown formatting
	content = regexp.MustCompile(`\*\*(.*?)\*\*`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`\*(.*?)\*`).ReplaceAllString(content, "$1")
	content = regexp.MustCompile(`\n\n`).ReplaceAllString(content, "\n")
	content = regexp.MustCompile(`\n- `).ReplaceAllString(content, "\n• ")

	// Trim whitespace
	content = strings.TrimSpace(content)

	// Limit to 280 characters (Twitter limit)
	// if len(content) > 280 {
	// 	// Try to cut at a sentence boundary
	// 	cutIndex := 220
	// 	for i := 277; i >= 10; i-- {
	// 		if content[i] == '.' || content[i] == '!' || content[i] == '?' {
	// 			cutIndex = i + 1
	// 			break
	// 		}
	// 	}
	// 	content = content[:cutIndex] + "..."
	// }

	return content
}
