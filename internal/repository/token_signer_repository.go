package repository

/*
TokenSigner is the contract for issuing an authentication token for a
given user. The usecase layer depends only on this interface, so the
underlying token strategy (JWT, opaque session tokens, etc.) can be
swapped without touching business logic.
*/
type TokenSigner interface {
	/*
		Sign generates a new token bound to the given user ID.
	*/
	Sign(userID string) (string, error)
}
