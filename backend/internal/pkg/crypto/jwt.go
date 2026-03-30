package crypto

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for a ReadBud user.
type Claims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration values.
type JWTConfig struct {
	Secret string
	Expiry time.Duration // e.g. 24 * time.Hour
}

// GenerateToken creates a signed JWT for the given user.
func GenerateToken(cfg JWTConfig, userID int64, username, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.Expiry)),
			Issuer:    "readbud",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(cfg.Secret))
	if err != nil {
		return "", fmt.Errorf("crypto.GenerateToken: %w", err)
	}
	return signed, nil
}

// ParseToken validates and parses a JWT string, returning the claims.
func ParseToken(cfg JWTConfig, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(cfg.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("crypto.ParseToken: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("crypto.ParseToken: invalid token claims")
	}

	return claims, nil
}

// RefreshToken generates a new token with a fresh expiry from existing claims.
func RefreshToken(cfg JWTConfig, oldClaims *Claims) (string, error) {
	return GenerateToken(cfg, oldClaims.UserID, oldClaims.Username, oldClaims.Role)
}
