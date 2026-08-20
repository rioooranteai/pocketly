package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

/*
JWTSigner implements repository.TokenSigner using signed JWTs.
Tokens are signed with HS256 and expire after ttl.
*/
type JWTSigner struct {
	secret []byte
	ttl    time.Duration
}

/*
NewJWTSigner builds a JWTSigner using the given secret key.
Tokens issued by this signer are valid for 24 hours.
*/
func NewJWTSigner(secret string) *JWTSigner {
	return &JWTSigner{
		secret: []byte(secret),
		ttl:    24 * time.Hour,
	}
}

/*
Sign creates a signed JWT bound to the given user ID. The token
carries the user ID as its subject claim, along with issued-at and
expiry timestamps.
*/
func (jw *JWTSigner) Sign(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(jw.ttl).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jw.secret)
}
