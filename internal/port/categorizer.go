package port

import (
	"context"
)

/*
Categorizer is the contract for automatically determining a
transaction's category from text describing it. The usecase layer
depends only on this interface, so the underlying strategy (a simple
keyword-matching text-based one, or a real AI provider like TypeSafe/OpenAI/
Gemini) can be swapped without touching business logic.
*/
type Categorizer interface {
	/*
		Categorize analyzes text describing a transaction and returns
		the most appropriate category, which should be one of the
		categories defined in the domain package. The usecase builds
		text from the description followed by every item name, so it
		may span several lines.
	*/
	Categorize(ctx context.Context, text string) (string, error)
}
