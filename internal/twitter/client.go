package twitter

import (
	"context"
	"fmt"

	"github.com/FinOwlX/internal/config"
	"github.com/michimani/gotwi"
	"github.com/michimani/gotwi/tweet/managetweet"
	"github.com/michimani/gotwi/tweet/managetweet/types"
)

// Client wraps the Twitter client
type Client struct {
	client *gotwi.Client
}

// NewClient creates a new Twitter client
func NewClient(cfg *config.Config) (*Client, error) {
	// Set up OAuth1 configuration for gotwi
	in := &gotwi.NewClientInput{
		AuthenticationMethod: gotwi.AuthenMethodOAuth1UserContext,
		APIKey:               cfg.APIKey,
		APIKeySecret:         cfg.APIKeySecret,
		OAuthToken:           cfg.OAuthToken,
		OAuthTokenSecret:     cfg.OAuthTokenSecret,
	}

	// Create Twitter client
	client, err := gotwi.NewClient(in)
	if err != nil {
		return nil, fmt.Errorf("failed to create Twitter client: %w", err)
	}

	return &Client{
		client: client,
	}, nil
}

// PostTweet posts a tweet with the given message
func (c *Client) PostTweet(text string) (string, error) {
	params := &types.CreateInput{
		Text: gotwi.String(text),
	}

	res, err := managetweet.Create(context.Background(), c.client, params)
	if err != nil {
		return "", fmt.Errorf("failed to post tweet: %w", err)
	}

	return gotwi.StringValue(res.Data.ID), nil
}

// DeleteTweet deletes a tweet specified by tweet ID
func (c *Client) DeleteTweet(id string) (bool, error) {
	params := &types.DeleteInput{
		ID: id,
	}

	res, err := managetweet.Delete(context.Background(), c.client, params)
	if err != nil {
		return false, fmt.Errorf("failed to delete tweet: %w", err)
	}

	return gotwi.BoolValue(res.Data.Deleted), nil
}
