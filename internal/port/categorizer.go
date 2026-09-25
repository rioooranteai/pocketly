package port

import (
	"context"
)

/*
Categorizer is the contract for automatically determining a
transaction's category from its description. The usecase layer
depends only on this interface, so the underlying strategy (a simple
keyword-matching dummy, or a real AI provider like TypeSafe/OpenAI/
Gemini) can be swapped without touching business logic.
*/
type Categorizer interface {
	/*
		Categorize analyzes a transaction description and returns the
		most appropriate category, which should be one of the categories
		defined in the domain package.
	*/
	Categorize(ctx context.Context, description string) (string, error)
}
