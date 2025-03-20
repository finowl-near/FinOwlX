package ai

import (
	"context"
	"errors"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

var (
	// ErrEnhanceContent is returned when the AI fails to enhance content
	ErrEnhanceContent = errors.New("AI failed to enhance content")
)

// Client represents an AI client for enhancing content
type Client struct {
	APIKey  string
	Model   string
	BaseURL string
}

// NewDeepSeekAI creates a new DeepSeek AI client
func NewDeepSeekAI(APIKey string) *Client {
	return &Client{
		APIKey:  APIKey,
		Model:   "deepseek-chat",
		BaseURL: "https://api.deepseek.com",
	}
}

// EnhanceContent enhances the given content using AI
func (ai *Client) EnhanceContent(ctx context.Context, content string) (string, error) {
	client := openai.NewClient(
		option.WithAPIKey(ai.APIKey),
		option.WithBaseURL(ai.BaseURL),
	)

	// Create a prompt specific to the section type
	prompt := ai.createPromptForSection()

	fmt.Println("--------------------------------------------------")

	fmt.Println(prompt)

	fmt.Println("--------------")

	fmt.Println(content)

	fmt.Println("--------------------------------------------------")

	chatCompletion, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(prompt),
			openai.UserMessage(content),
		}),
		Model: openai.F(ai.Model),
	})
	if err != nil {
		return "", fmt.Errorf("%w: %w", ErrEnhanceContent, err)
	}

	if len(chatCompletion.Choices) < 1 {
		return "", fmt.Errorf("%w: empty response", ErrEnhanceContent)
	}

	return chatCompletion.Choices[0].Message.Content, nil
}

// createPromptForSection creates a specific prompt based on the section type
func (ai *Client) createPromptForSection() string {
	return `Act as a professional crypto analyst and Twitter growth expert. Your goal is to transform raw crypto market insights into highly engaging, viral Twitter posts. The tone should be authoritative, insightful, and engaging, with a perfect balance of professionalism and hype:
	•	Introduce the most exciting trend of the day in an eye-catching way.
	•	Explain why each token/project is trending, referencing key catalysts like institutional moves, ETF approvals, on-chain activity, and major influencer sentiment.
	•	End with a line crediting @finowl_finance as the data provider.

	### **Rules for Generation:**  
	1. **Only use the tokens I provided**. Do not add any extra ones.  
	2. **Only use the reasons I provided**. Do not create new trends.  
	3. **Follow the exact structure and format given.**  
	4. **Use an authoritative yet engaging tone.**  

	Now, generate a Twitter post using the exact information`
}
