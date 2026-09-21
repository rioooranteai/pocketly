package ai

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"

	"pocketly/internal/domain"
	"pocketly/internal/infrastructure/config"
)

/*
OpenAIVisionExtractor implements repository.VisionExtractor using
OpenAI's vision-capable chat completion API. It sends the receipt
image as a base64 data URL alongside an instruction asking for a
strict JSON response, then parses that JSON into domain entities.
*/
type OpenAIVisionExtractor struct {
	client      openai.Client
	maxFileSize int64
}

/*
NewOpenAIVisionExtractor builds an OpenAIVisionExtractor using the
given vision-specific configuration (API key and max upload size).
*/
func NewOpenAIVisionExtractor(cfg config.VisionConfig) *OpenAIVisionExtractor {
	client := openai.NewClient(option.WithAPIKey(cfg.APIKey))
	return &OpenAIVisionExtractor{client: client, maxFileSize: cfg.MaxFileSize}
}

/*
extractedItem is the intermediate shape used to unmarshal a single
line item from the model's JSON response, before it is converted
into a domain.TransactionItem (which requires an ID and a
TransactionID that the model has no business generating).
*/
type extractedItem struct {
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

/*
extractedReceipt is the intermediate shape used to unmarshal the
model's full JSON response before mapping it to domain types.
*/
type extractedReceipt struct {
	Description string          `json:"description"`
	Items       []extractedItem `json:"items"`
}

/*
visionPrompt instructs the model to return a strictly-formatted JSON
object matching extractedReceipt, with no additional commentary.
*/
const visionPrompt = `You are given an image of a shopping receipt.
Extract the following information and respond with ONLY a valid JSON object,
no other text, no markdown code fences, matching exactly this shape:

{
  "description": "short summary of the merchant or transaction",
  "items": [
    {"name": "item name", "quantity": 1, "price": 15000}
  ]
}

If a field cannot be determined, use a reasonable default (quantity 1, price 0).`

/*
Extract validates the image size, sends it to OpenAI's vision model
along with visionPrompt, and parses the resulting JSON into a
description and a list of domain.TransactionItem. Each item is
assigned a fresh UUID; TransactionID is left empty since the caller
(TransactionUsecase) assigns it once the parent transaction exists.
*/
func (o *OpenAIVisionExtractor) Extract(ctx context.Context, imageData []byte) (string, []domain.TransactionItem, error) {
	if len(imageData) == 0 {
		return "", nil, domain.ErrEmptyImageData
	}
	if int64(len(imageData)) > o.maxFileSize {
		return "", nil, domain.ErrImageSizeExceedsLimit
	}

	encoded := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:image/jpeg;base64,%s", encoded)

	completion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4o,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(visionPrompt),
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: dataURL,
				}),
			}),
		},
	})
	if err != nil {
		return "", nil, fmt.Errorf("openai vision request failed: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", nil, fmt.Errorf("openai vision returned no choices")
	}

	rawContent := completion.Choices[0].Message.Content

	var parsed extractedReceipt
	if err := json.Unmarshal([]byte(rawContent), &parsed); err != nil {
		return "", nil, fmt.Errorf("failed to parse vision response as JSON: %w", err)
	}

	items := make([]domain.TransactionItem, 0, len(parsed.Items))
	for _, item := range parsed.Items {
		items = append(items, domain.TransactionItem{
			ID:       uuid.New().String(),
			Name:     item.Name,
			Quantity: item.Quantity,
			Price:    item.Price,
		})
	}

	return parsed.Description, items, nil
}
