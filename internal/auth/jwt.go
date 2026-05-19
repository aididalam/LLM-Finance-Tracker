package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 30 * 24 * time.Hour

type Claims struct {
	ChatID int64  `json:"chat_id"`
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func IssueToken(chatID int64, userID, secret string) (string, error) {
	claims := Claims{
		ChatID: chatID,
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
