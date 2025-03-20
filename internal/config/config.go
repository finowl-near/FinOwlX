package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	APIKeyEnvKeyName           = "GOTWI_API_KEY"
	APIKeySecretEnvKeyName     = "GOTWI_API_KEY_SECRET"
	OAuthTokenEnvKeyName       = "GOTWI_ACCESS_TOKEN"
	OAuthTokenSecretEnvKeyName = "GOTWI_ACCESS_TOKEN_SECRET"
	DefaultTweetTextEnvName    = "DEFAULT_TWEET_TEXT"
	FinowlStartIDEnvName       = "FINOWL_START_ID"
	DeepSeekAPIKeyEnvName      = "DEEPSEEK_API_KEY"
)

// Config holds all configuration for the application
type Config struct {
	APIKey           string
	APIKeySecret     string
	OAuthToken       string
	OAuthTokenSecret string
	DefaultTweetText string
	FinowlStartID    int
	DeepSeekAPIKey   string
}

// Load loads the configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	config := &Config{
		APIKey:           os.Getenv(APIKeyEnvKeyName),
		APIKeySecret:     os.Getenv(APIKeySecretEnvKeyName),
		OAuthToken:       os.Getenv(OAuthTokenEnvKeyName),
		OAuthTokenSecret: os.Getenv(OAuthTokenSecretEnvKeyName),
		DefaultTweetText: os.Getenv(DefaultTweetTextEnvName),
		DeepSeekAPIKey:   os.Getenv(DeepSeekAPIKeyEnvName),
	}

	// Parse Finowl start ID
	startIDStr := os.Getenv(FinowlStartIDEnvName)
	if startIDStr != "" {
		startID, err := strconv.Atoi(startIDStr)
		if err != nil {
			return nil, errors.New("invalid FINOWL_START_ID: must be a number")
		}
		config.FinowlStartID = startID
	} else {
		// Default to ID 105 if not specified
		config.FinowlStartID = 105
	}

	// Validate required fields
	if config.APIKey == "" || config.APIKeySecret == "" {
		return nil, errors.New("missing required API credentials in environment variables")
	}
	if config.OAuthToken == "" || config.OAuthTokenSecret == "" {
		return nil, errors.New("missing required OAuth tokens in environment variables")
	}

	// Set default tweet text if not provided
	if config.DefaultTweetText == "" {
		config.DefaultTweetText = "This is an automated tweet from my Go application!"
	}

	return config, nil
}
