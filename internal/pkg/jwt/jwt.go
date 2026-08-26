package jwt

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type Key struct {
	SecretKey string
}

func NewJWTToken(config *viper.Viper) *Key {
	sekretKey := config.GetString("jwt.secret_key")
	return &Key{
		SecretKey: sekretKey,
	}
}

func (key *Key) GenerateToken(userId, email, role string, duration time.Duration) (string, *UserClaims, error) {
	claims, err := NewUserClaims(userId, email, role, duration)
	if err != nil {
		return "", nil, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := token.SignedString([]byte(key.SecretKey))
	if err != nil {
		return "", nil, err
	}

	return signedString, claims, nil
}

func (key *Key) VerifyToken(tokenString string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, fiber.NewError(fiber.StatusInternalServerError, "Unexpected sign methods")
		}

		return []byte(key.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if ok && token.Valid {
		return claims, nil
	}

	return nil, fiber.NewError(fiber.StatusInternalServerError, "invalid token")
}
