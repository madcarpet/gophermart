package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/madcarpet/gophermart/internal/logger"
	"go.uber.org/zap"
)

type jwtTokenizer struct {
	secretKey      string
	expirationTime time.Duration
}

type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

func NewJwtTokenizer(key string, etime time.Duration) *jwtTokenizer {
	return &jwtTokenizer{secretKey: key, expirationTime: etime}
}

func (t *jwtTokenizer) ProduceToken(id string) (string, error) {
	// Token producing with expiration time and user id
	uid := id
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(t.expirationTime)),
		},
		UserID: uid,
	})
	tokenString, err := token.SignedString([]byte(t.secretKey))
	if err != nil {
		logger.Log.Error("token generating error", zap.Error(err))
		return "", err
	}
	return tokenString, nil
}

func (t *jwtTokenizer) VerifyToken(ts string) (string, error) {
	secretKey := t.secretKey
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(ts, claims,
		func(tn *jwt.Token) (interface{}, error) {
			//Check if the token signed with HMAC
			if _, ok := tn.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", tn.Header["alg"])
			}
			return []byte(secretKey), nil
		})
	if err != nil {
		logger.Log.Error("token validating error", zap.Error(err))
		return "", err
	}
	if !token.Valid {
		logger.Log.Error("token is invalid")
		return "", fmt.Errorf("token %s is invalid", ts)
	}
	return claims.UserID, nil
}
