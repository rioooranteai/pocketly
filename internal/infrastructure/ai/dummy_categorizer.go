package ai

import (
	"context"
	"strings"

	"pocketly/internal/domain"
	"pocketly/internal/port"
)

/*
DummyCategorizer is a rule-based stand-in for a real AI categorizer.
It exists so the rest of the app (usecase, handler, tests) can depend
on the Categorizer port without needing a real LLM call — useful for
local development, unit tests, and before the AI integration
(OpenAI/Claude/local model) is wired in.
*/
type DummyCategorizer struct{}

var _ port.Categorizer = (*DummyCategorizer)(nil)

/*
NewDummyCategorizer builds a DummyCategorizer. It takes no
dependencies because it does not call any external service.
*/
func NewDummyCategorizer() *DummyCategorizer {
	return &DummyCategorizer{}
}

/*
categoryKeywords maps a category name to substrings that, if found
(case-insensitive) in a transaction description, are considered a
match. This is intentionally simple — just enough to make local
testing and demos behave sensibly without calling a real AI model.
*/
var categoryKeywords = map[string][]string{
	domain.CategoryFood:           {"makan", "resto", "warung", "kopi", "cafe", "nasi", "ayam"},
	domain.CategoryTransportation: {"bensin", "grab", "gojek", "ojol", "parkir", "tol", "pertamina"},
	domain.CategoryShopping:       {"belanja", "shopee", "tokopedia", "mall", "baju"},
	domain.CategoryBills:          {"listrik", "pulsa", "wifi", "internet", "pdam", "bpjs"},
	domain.CategoryEntertainment:  {"nonton", "bioskop", "netflix", "spotify", "game"},
	domain.CategoryHealth:         {"apotek", "obat", "dokter", "rumah sakit", "klinik"},
}

/*
Categorize assigns a category based on simple keyword matching
against the transaction description. It returns
domain.CategoryUncategorized as a safe default when nothing matches,
and never returns an error — this dummy implementation has no
external dependency that can fail.
*/
func (d *DummyCategorizer) Categorize(ctx context.Context, description string) (string, error) {
	lowered := strings.ToLower(description)

	for category, keywords := range categoryKeywords {
		for _, kw := range keywords {
			if strings.Contains(lowered, kw) {
				return category, nil
			}
		}
	}

	return domain.CategoryUncategorized, nil
}
