package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	domain "milpa/domain/entities"
	port "milpa/domain/port/secondary"
)

type JWTProvider struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTProvider(secret string, expiration time.Duration) *JWTProvider {
	if expiration <= 0 {
		expiration = 24 * time.Hour
	}
	return &JWTProvider{
		secret:     []byte(secret),
		expiration: expiration,
	}
}

type customClaims struct {
	UserID string             `json:"user_id"`
	Role   domain.RoleOptions `json:"role"`
	jwt.RegisteredClaims
}

func (p *JWTProvider) GenerateToken(userID uuid.UUID, role domain.RoleOptions) (string, error) {
	now := time.Now()

	claims := customClaims{
		UserID: userID.String(),
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(p.expiration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString(p.secret)
	if err != nil {
		return "", fmt.Errorf("jwt.GenerateToken: %w", err)
	}

	return signed, nil
}

func (p *JWTProvider) ValidateToken(raw string) (*port.JWTClaims, error) {
	claims := &customClaims{}

	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return p.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("jwt.ValidateToken: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("jwt.ValidateToken: %w", domain.ErrUnauthorized)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("jwt.ValidateToken: invalid user_id claim: %w", err)
	}

	return &port.JWTClaims{
		UserID: userID,
		Role:   claims.Role,
	}, nil
}

var _ port.JWTProvider = (*JWTProvider)(nil)
