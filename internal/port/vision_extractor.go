package port

import (
	"context"

	"pocketly/internal/domain"
)

/*
VisionExtractor is the contract for reading a receipt image and
returning its description and line items. The usecase layer depends
only on this interface, so the vision provider (OpenAI, Gemini, a
local OCR model) can be swapped without touching business logic.
Size and item validation are business rules and stay in the usecase,
not in implementations of this interface.
*/
type VisionExtractor interface {
	/*
		Extract reads the receipt in imageData and returns a short
		description of the merchant or transaction plus its line items.
		Items come back without ID or TransactionID; the usecase assigns
		both. imageData has already been checked for emptiness and size
		by the caller. An error means the provider call or its response
		failed, not that the receipt had no items.
	*/
	Extract(ctx context.Context, imageData []byte) (description string, items []domain.TransactionItem, err error)
}
