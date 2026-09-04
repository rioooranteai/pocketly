package ai

import "context"

type DummyCategorizer struct{}

func NewDummyCategorizer() {

}

func Categorize(ctx context.Context, description string) (string, error) {
	return "uncategorized", nil
}

