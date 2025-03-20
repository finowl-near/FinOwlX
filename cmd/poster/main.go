package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/FinOwlX/internal/ai"
	"github.com/FinOwlX/internal/config"
	"github.com/FinOwlX/internal/finowl"
	"github.com/FinOwlX/internal/twitter"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("Starting X poster application")

	// Define command line flags
	useFinowl := flag.Bool("finowl", false, "Use Finowl API to post market summaries")
	manualTweet := flag.String("tweet", "", "Post a manual tweet with the given text")
	disableAI := flag.Bool("no-ai", false, "Disable AI enhancement of tweets")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create Twitter client
	twitterClient, err := twitter.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Twitter client: %v", err)
	}

	// Create AI client if API key is available and AI is not disabled
	var aiClient *ai.Client
	if cfg.DeepSeekAPIKey != "" && !*disableAI {
		aiClient = ai.NewDeepSeekAI(cfg.DeepSeekAPIKey)
		log.Println("AI enhancement enabled")
	} else if *disableAI {
		log.Println("AI enhancement disabled by flag")
	} else if cfg.DeepSeekAPIKey == "" {
		log.Println("AI enhancement disabled: No DeepSeek API key provided")
	}

	// If using Finowl mode
	if *useFinowl {
		log.Println("Starting in Finowl mode")
		finowlService := finowl.NewService(twitterClient, cfg.FinowlStartID, aiClient)
		finowlService.RunContinuously()
		return
	}

	// If a manual tweet was specified
	if *manualTweet != "" {
		postManualTweet(twitterClient, *manualTweet)
		return
	}

	// Otherwise, use command line args or default message
	message := cfg.DefaultTweetText
	if len(flag.Args()) > 0 {
		message = flag.Args()[0]
	}

	postManualTweet(twitterClient, message)
}

func postManualTweet(client *twitter.Client, message string) {
	// Post tweet
	tweetID, err := client.PostTweet(message)
	if err != nil {
		log.Fatalf("Failed to post tweet: %v", err)
	}

	// Display success message
	fmt.Printf("Successfully posted tweet with ID: %s\n", tweetID)
	fmt.Printf("View at: https://twitter.com/user/status/%s\n", tweetID)
}
