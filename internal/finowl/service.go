package finowl

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/FinOwlX/internal/ai"
	"github.com/FinOwlX/internal/twitter"
	"golang.org/x/exp/rand"
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
	// Initialize rate limit
	remainingRateLimit := 17 // Total rate limit available

	// Check if we can post segments first
	if remainingRateLimit > 6 { // Ensure we leave 6 for future summaries

		ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
		defer cancel()

		prompt := s.aiClient.CreatePromptForSectionSegements()

		enhancedContent, err := s.aiClient.EnhanceContent(ctx, content, prompt)
		if err != nil {
			log.Printf("Warning: Failed to enhance content with AI: %v. Using original content.", err)
		} else {
			content = cleanTickers(enhancedContent)
			log.Printf("Successfully enhanced content with AI")
		}
		// Decide whether to post segments or the full summary first
		segments := twitter.SplitCryptoTweet(content)
		segmentsToPost := len(segments) - 1 // Skip first segment

		for i := 1; i <= segmentsToPost; i++ {
			cleanSegment := removeAsterisks(segments[i])
			sleepDuration := time.Duration(600+rand.Intn(1000)) * time.Second

			segmentTweetID, err := s.twitterClient.PostTweet(cleanSegment)
			if err != nil {
				log.Printf("Warning: Failed to post segment %d: %v", i, err)
				break // Stop posting segments if we hit an error
			}
			log.Printf("Posted segment %d with ID: %s", i, segmentTweetID)

			remainingRateLimit--         // Decrement rate limit for each successful post
			if remainingRateLimit <= 6 { // Check if we need to stop posting segments
				log.Printf("Reached limit for segments, stopping to preserve rate limit for summaries.")
				break
			}

			time.Sleep(sleepDuration)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Second)
	defer cancel()

	prompt := s.aiClient.CreatePromptForSection()

	enhancedContent, err := s.aiClient.EnhanceContent(ctx, content, prompt)
	if err != nil {
		log.Printf("Warning: Failed to enhance content with AI: %v. Using original content.", err)
	} else {
		content = cleanTickers(enhancedContent)
		log.Printf("Successfully enhanced content with AI")
	}

	// First post the full content
	fmt.Println("Posting full content:")
	fmt.Println("=====================================================")

	fmt.Println(content)
	fmt.Println("=====================================================")

	_, err = s.twitterClient.PostTweet(content)
	if err != nil {
		log.Printf("Warning: Failed to post content : %v", err)
	}
	log.Printf("Posted content succefully  ...")

	remainingRateLimit--         // Decrement rate limit for each successful post
	if remainingRateLimit == 0 { // Check if we need to stop posting segments
		log.Printf("Reached limit for everything .....")
	}

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

		// Wait 2 hours before checking for the next summary
		log.Printf("Successfully posted summary ID %d. Waiting 4 hours for next summary...", s.currentID-1)
		time.Sleep(2 * time.Hour)
	}
}

// // postCryptoTweets takes an array of tweet segments, skips the first & last, and posts the valid ones
// func postCryptoTweets(client *twitter.Client, segments []string) {
// 	// Ensure we have enough segments to process
// 	if len(segments) <= 2 {
// 		fmt.Println("Not enough content to post.")
// 		return
// 	}

// 	// Iterate over the valid segments (skip first and last)
// 	for i := 1; i < len(segments); i++ {
// 		cleanSegment := removeAsterisks(segments[i])
// 		sleepDuration := time.Duration(600+rand.Intn(1000)) * time.Second

// 		tweet := cleanSegment // + "\n\n📊 Data powered by @finowl_finance #crypto"
// 		segmentTweetID, err := client.PostTweet(tweet)
// 		if err != nil {
// 			log.Printf("Warning: Failed to post segment %d: %v", i+1, err)
// 			time.Sleep(sleepDuration)
// 			continue // Continue with next segment even if this one fails
// 		}
// 		log.Printf("Posted segment %d/%d with ID: %s", i+1, len(segments), segmentTweetID)

// 		time.Sleep(sleepDuration)

// 	}
// }

// removeAsterisks removes all ** from the content and adds space after opening parentheses
func removeAsterisks(content string) string {
	// Define regex pattern to match ** and replace with empty string
	re := regexp.MustCompile(`\*\*`)
	content = re.ReplaceAllString(content, "")

	// Add space after opening parentheses
	reParens := regexp.MustCompile(`\(`)
	content = reParens.ReplaceAllString(content, "( ")

	return content
}
