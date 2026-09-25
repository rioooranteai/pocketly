package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"pocketly/internal/repository"
)

/*
Defaults and tuning values for TypeSafeCategorizer. The base URL,
model, and timeout mirror the defaults of TypeSafe's official SDKs.
typeSafeMinConfidence is the lowest confidence accepted before the
answer is treated as a guess and replaced with "uncategorized".
*/
const (
	typeSafeDefaultBaseURL = "https://api.typesafe.ai"
	typeSafeDefaultModel   = "jev-latest"
	typeSafeTimeout        = 10 * time.Second
	typeSafeMinConfidence  = 0.6
	typeSafeQuestionID     = "category"
	typeSafeUncategorized  = "uncategorized"
	typeSafeErrorBodyLimit = 4096
)

/*
typeSafeCategoryCriteria lists the categories Jev may choose from,
each with a description the model uses to decide. The names match
DummyCategorizer's so stored categories stay consistent whichever
implementation is wired in.
*/
var typeSafeCategoryCriteria = map[string]string{
	"food":           "Makanan, minuman, restoran, warung, kopi, belanja bahan dapur",
	"transportation": "Bensin, ojek online, taksi, parkir, tol, tiket kendaraan umum",
	"shopping":       "Belanja barang non-makanan: pakaian, elektronik, marketplace, mall",
	"bills":          "Tagihan rutin: listrik, air, internet, pulsa, BPJS, cicilan",
	"entertainment":  "Hiburan: bioskop, game, langganan streaming musik atau film",
	"health":         "Kesehatan: apotek, obat, dokter, klinik, rumah sakit",
}

/*
typeSafeQuestion and typeSafeRequest are the request body shape of
TypeSafe's POST /v1/systemone endpoint.
*/
type typeSafeQuestion struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria,omitempty"`
}

type typeSafeRequest struct {
	State     string                      `json:"state"`
	Model     string                      `json:"model"`
	Questions map[string]typeSafeQuestion `json:"questions"`
}

/*
typeSafeAnswer and typeSafeResponse hold only the response fields a
choice question needs; the rest of the payload is ignored.
*/
type typeSafeAnswer struct {
	Type       string  `json:"type"`
	Choice     string  `json:"choice"`
	Confidence float64 `json:"confidence"`
}

type typeSafeResponse struct {
	Answers map[string]typeSafeAnswer `json:"answers"`
}

type TypeSafeCategorizer struct {
	apyKey     string
	httpClient *http.Client
	baseUrl string
	model string
}

var _ repository.CategorizerRepository = (*TypeSafeCategorizer)(nil)

func NewTypeSafeCategorizer(apiKey string, baseUrl string, model string) *TypeSafeCategorizer {
	if baseUrl == "" {
		baseUrl = typeSafeDefaultBaseURL
	}
	if model == "" {
		model = typeSafeDefaultModel
	}

	return &TypeSafeCategorizer{
		apyKey: apiKey,
		httpClient: &http.Client{Timeout: typeSafeTimeout},
		baseUrl: baseUrl,
		model: model,
	}
}

/*
Categorize asks Jev to pick one category from typeSafeCategoryCriteria
for the given description. Technical failures (network, non-200
status, malformed or unexpected response) are returned as errors so
the usecase logs them and falls back. An empty description or a
low-confidence answer is not a failure, so it returns "uncategorized"
with no error.
*/
func (t *TypeSafeCategorizer) Categorize(ctx context.Context, description string) (string, error) {
	if strings.TrimSpace(description) == "" {
		return typeSafeUncategorized, nil
	}
	if t.apyKey == "" {
		return "", errors.New("typesafe API key is not configured")
	}

	body, err := json.Marshal(typeSafeRequest{
		State: description,
		Model: t.model,
		Questions: map[string]typeSafeQuestion{
			typeSafeQuestionID: {
				Type:         "choice",
				Instructions: "Which spending category best fits this transaction?",
				Criteria:     typeSafeCategoryCriteria,
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to encode typesafe request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(t.baseUrl, "/")+"/v1/systemone", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to build typesafe request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+t.apyKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("typesafe request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, typeSafeErrorBodyLimit))
		return "", fmt.Errorf("typesafe returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	var parsed typeSafeResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("failed to parse typesafe response as JSON: %w", err)
	}

	answer, ok := parsed.Answers[typeSafeQuestionID]
	if !ok {
		return "", fmt.Errorf("typesafe response has no %q answer", typeSafeQuestionID)
	}
	if _, known := typeSafeCategoryCriteria[answer.Choice]; !known {
		return "", fmt.Errorf("typesafe returned unknown category %q", answer.Choice)
	}
	if answer.Confidence < typeSafeMinConfidence {
		return typeSafeUncategorized, nil
	}

	return answer.Choice, nil
}