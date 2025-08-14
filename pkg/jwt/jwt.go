package jwt

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"gophemart/pkg/logger"
	"time"
)

type Manager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

func NewManager(secret string, duration time.Duration) *Manager {
	return &Manager{
		secretKey:     []byte(secret),
		tokenDuration: duration,
	}
}

func (m *Manager) GenerateToken(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(m.tokenDuration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secretKey)
}

func (m *Manager) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	logger.Debug().
		Msg("Validating JWT token")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			err := fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			logger.Warn().
				Err(err).
				Str("algorithm", fmt.Sprintf("%v", token.Header["alg"])).
				Msg("Invalid token signing method")
			return nil, err
		}
		return m.secretKey, nil
	})
	if err != nil {
		logger.Error().
			Err(err).
			Msg("Unexpected token validation error")
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, _ := claims["user_id"].(float64)
		exp, _ := claims["exp"].(float64)
		logger.Debug().
			Uint("user_id", uint(userID)).
			Time("expires_at", time.Unix(int64(exp), 0)).
			Msg("JWT token validated successfully")
		return claims, nil
	}
	logger.Warn().
		Msg("Invalid token claims")

	return nil, fmt.Errorf("invalid token")
}
