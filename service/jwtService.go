package service

import (
	"mofe64/playlistGen/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	GenerateToken(username string) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtService struct {
	secret string
}

func NewJWTService() JWTService {
	return &jwtService{
		secret: config.JWTSecret(),
	}
}

func (j *jwtService) GenerateToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 24),
	})
	tokenString, err := token.SignedString(j.secret)
	if err != nil {
		return "", err
	}
	return tokenString, err
}

func (j *jwtService) ValidateToken(token string) (*jwt.Token, error) {
	return nil, nil
}
