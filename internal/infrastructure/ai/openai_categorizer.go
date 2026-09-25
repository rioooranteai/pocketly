package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"pocketly/internal/port"
)

const categorizerTimeout = 25 * time.Second

type extractedCategory struct {
	Category string `json:"category"`
}

type OpenAICategorizer struct {
	client openai.Client
}

var _ port.Categorizer = (*OpenAICategorizer)(nil)

func NewOpenAICategorizer(apiKey string) *OpenAIVisionExtractor {
	client := openai.NewClient(option.WithAPIKey(apiKey))
	return &OpenAIVisionExtractor{client: client}
}

const categorizerPrompt = `You are a financial transaction classification system.
Your task: categorize the given receipt/transaction description into EXACTLY ONE of the following categories:

- food: Food, drinks, restaurants, food stalls, coffee, grocery shopping
- transportation: Gas/fuel, ride-hailing, taxi, parking, tolls, public transit tickets
- shopping: Non-food shopping: clothing, electronics, marketplace, mall
- bills: Recurring bills: electricity, water, internet, phone credit, BPJS, installments
- entertainment: Entertainment: cinema, games, music or video streaming subscriptions
- health: Health: pharmacy, medicine, doctor, clinic, hospital

If the description doesn't clearly match any category, choose the closest one based on general context.

Respond with ONLY valid JSON, no other text, no markdown code fences, matching exactly this shape:

{
  "category": "one of: food, transportation, shopping, bills, entertainment, health"
}`

func (o *OpenAICategorizer) Categorize(ctx context.Context, description string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, visionTimeout)
	defer cancel()

	completion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT5_6Luna,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(categorizerPrompt),
			openai.UserMessage(description),
		},
	})

	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("openai returned no category")
	}

	rawContent := completion.Choices[0].Message.Content

	var parsed extractedCategory
	if err := json.Unmarshal([]byte(rawContent), &parsed); err != nil {
		return "", fmt.Errorf("failed to parse llm response as JSON: %w", err)
	}

	return parsed.Category, nil
}
