package repository

import (
	"context"
)

/*
CategorizerRepository is the contract for automatically determining a
transaction's category from its description. The usecase layer
depends only on this interface, so the underlying strategy (a simple
keyword-matching dummy, or a real AI provider like Claude/OpenAI/
Gemini) can be swapped without touching business logic.
*/
type CategorizerRepository interface {
	/*
		Categorize analyzes a transaction description and returns the
		most appropriate category.
	*/
	Categorize(ctx context.Context, description string) (string, error)
}
