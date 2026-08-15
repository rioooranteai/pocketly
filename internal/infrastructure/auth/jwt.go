package auth

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
)

type JWTSigner struct {
	secret []byte
	ttl    time.Duration
}

func NewJWTSigner(secret string) *JWTSigner {
	return &JWTSigner{
		secret: []byte(secret),
		ttl: 24 * time.Hour,
	}
}

func (jw *JWTSigner) Sign(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp" : time.Now().Add(jw.ttl).Unix(),
		"iat" : time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(jw.secret)

}