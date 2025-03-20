package finowl

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/FinOwlX/internal/ai"
	"github.com/FinOwlX/internal/twitter"
)

// Service manages the process of fetching summaries and posting to Twitter
type Service struct {
	finowlClient  *Client
	twitterClient *twitter.Client
	aiClient      *ai.Client
	currentID     int
	useAI         bool
}

// NewService creates a new Finowl service
func NewService(twitterClient *twitter.Client, startID int, aiClient *ai.Client) *Service {
	return &Service{
		finowlClient:  NewClient(),
		twitterClient: twitterClient,
		aiClient:      aiClient,
		currentID:     startID,
		useAI:         aiClient != nil,
	}
}

// PostLatestSummary fetches the latest summary and posts it to Twitter
func (s *Service) PostLatestSummary() error {
	// Get the current summary
	summary, err := s.finowlClient.GetSummary(s.currentID)
	if err != nil {
		return err
	}

	// Parse the content
	sections, err := s.finowlClient.ParseContent(summary.Summary.Content)
	if err != nil {
		return err
	}

	// Post each section to Twitter
	fmt.Println("============")
	fmt.Println(sections.FeaturedTickers)
	fmt.Println("============")

	err = s.postSection(sections.FeaturedTickers)
	if err != nil {
		return err
	}

	// Wait a bit between tweets to avoid rate limiting
	time.Sleep(1 * time.Minute)

	// Update the current ID
	s.currentID = summary.Summary.ID + 1

	return nil
}
func cleanTickers(content string) string {
	// Define regex pattern to match **$TICKER** and preserve the $ sign
	re := regexp.MustCompile(`\*\*(\$\w+)\*\*`)
	// Replace with " $TICKER " ensuring the $ is included
	return re.ReplaceAllString(content, " $1 ")
}

// postSection posts a specific section to Twitter
func (s *Service) postSection(content string) error {
	sectionName := "Featured Tickers and Projects"
	// Enhance content with AI if available
	if s.useAI {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		log.Printf("Enhancing %s content with AI...", sectionName)
		enhancedContent, err := s.aiClient.EnhanceContent(ctx, content)
		if err != nil {
			log.Printf("Warning: Failed to enhance content with AI: %v. Using original content.", err)
		} else {
			content = cleanTickers(enhancedContent)
			log.Printf("Successfully enhanced content with AI")
		}
	}

	tweet := content

	fmt.Println(tweet)

	tweetID, err := s.twitterClient.PostTweet(tweet)
	if err != nil {
		return ErrTwitterPostFailed{Section: sectionName, Cause: err}
	}

	log.Printf("Posted %s tweet with ID: %s", sectionName, tweetID)
	return nil
}

// RunContinuously continuously fetches and posts summaries
func (s *Service) RunContinuously() {
	for {
		log.Printf("Processing summary ID: %d", s.currentID)

		err := s.PostLatestSummary()
		if err != nil {
			log.Printf("Error processing summary: %v", err)

			// Check if the error is because the summary doesn't exist yet
			if _, ok := err.(ErrSummaryNotFound); ok && s.currentID > 0 {
				log.Printf("Waiting for summary ID %d to become available...", s.currentID)
				summary, err := s.finowlClient.WaitForNextSummary(s.currentID - 1)
				if err != nil {
					log.Printf("Error waiting for next summary: %v", err)
					time.Sleep(15 * time.Minute)
					continue
				}

				s.currentID = summary.Summary.ID
				continue
			}

			// For other errors, wait a bit and try again
			log.Printf("Unexpected error, waiting 15 minutes before retrying...")
			time.Sleep(15 * time.Minute)
			continue
		}

		// Wait 4 hours before checking for the next summary
		log.Printf("Successfully posted summary ID %d. Waiting 4 hours for next summary...", s.currentID-1)
		time.Sleep(4 * time.Hour)
	}
}
