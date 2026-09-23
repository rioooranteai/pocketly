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

/*
TokenVerifier is the contract for validating an authentication token
and extracting the user it was issued for. The HTTP middleware depends
only on this interface, mirroring how the usecase layer depends on
TokenSigner.
*/
type TokenVerifier interface {
	/*
		ParseUserID validates the token and returns its user ID, or an
		error if the token is malformed, expired, or not trusted.
	*/
	ParseUserID(token string) (string, error)
}
