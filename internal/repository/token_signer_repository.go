package repository

type TokenSigner interface {
	Sign(userID string) (string, error)
}