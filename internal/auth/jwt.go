package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTService struct {
	config JWTConfig
}

func NewJWTService(config JWTConfig) *JWTService {
	return &JWTService{config: config}
}

func (j *JWTService) GenerateToken(user *User) (string, error) {

	claims := Claims{
		UserID: fmt.Sprintf("%d", user.ID),
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.config.Expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.config.Secret))
}

func (j *JWTService) ValidateToken(tokenStr string) (*Claims, error) {

	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, err
}
