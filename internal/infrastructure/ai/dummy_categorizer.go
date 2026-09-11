package ai

import (
	"context"
	"strings"
)

/*
DummyCategorizer is a rule-based stand-in for a real AI categorizer.
It exists so the rest of the app (usecase, handler, tests) can depend
on the Categorizer port without needing a real LLM call — useful for
local development, unit tests, and before the AI integration
(OpenAI/Claude/local model) is wired in.

It satisfies whatever Categorizer interface is defined in your
port/repository package. Uncomment and adjust the assertion below
once you confirm the exact interface name and import path.
*/
type DummyCategorizer struct{}

// var _ port.Categorizer = (*DummyCategorizer)(nil)

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
	"food": {"makan", "resto", "warung", "kopi", "cafe", "nasi", "ayam"},
	"transportation": {"bensin", "grab", "gojek", "ojol", "parkir", "tol", "pertamina"},
	"shopping": {"belanja", "shopee", "tokopedia", "mall", "baju"},
	"bills": {"listrik", "pulsa", "wifi", "internet", "pdam", "bpjs"},
	"entertainment": {"nonton", "bioskop", "netflix", "spotify", "game"},
	"health": {"apotek", "obat", "dokter", "rumah sakit", "klinik"},
}

/*
Categorize assigns a category based on simple keyword matching
against the transaction description. It returns "uncategorized" as
a safe default when nothing matches, and never returns an error —
this dummy implementation has no external dependency that can fail.
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

	return "uncategorized", nil
}
